package member

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(m Member) (Member, error) {
	if m.Name == "" {
		return Member{}, errors.New("name is required")
	}

	if m.Status == "" {
		m.Status = Active
	}

	if m.Status != Active && m.Status != Inactive {
		return Member{}, errors.New("invalid status")
	}

	now := time.Now()

	m.CreatedAt = now
	m.UpdatedAt = now

	result, err := s.repo.Create(m)
	if err != nil {
		return Member{}, err
	}

	return result, nil
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func (s *Service) Find(filters map[string]string, page, limit int) ([]Member, PaginationMeta, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	data, err := s.repo.Find(filters, limit, offset)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	total, err := s.repo.Count(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return data, meta, nil
}

func (s *Service) Update(id int64, payload Member) (Member, error) {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return Member{}, errors.New("member not found")
	}

	if payload.Name != "" {
		member.Name = payload.Name
	}

	if payload.Email != "" {
		member.Email = payload.Email
	}

	if payload.Phone != "" {
		member.Phone = payload.Phone
	}

	if payload.Status != "" {
		member.Status = payload.Status
	}

	if payload.ChurchRole != "" {
		member.ChurchRole = payload.ChurchRole
	}

	if payload.MaritalStatus != "" {
		member.MaritalStatus = payload.MaritalStatus
	}

	if payload.OriginDenomination != "" {
		member.OriginDenomination = payload.OriginDenomination
	}

	member.UpdatedAt = time.Now()

	return s.repo.Update(member)
}

func (s *Service) SoftDelete(id int64) error {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("member not found")
	}

	if member.Status == Inactive {
		return errors.New("member already inactive")
	}

	return s.repo.SoftDelete(id)
}

func (s *Service) ImportCSV(file io.Reader, userID int64) (ImportResult, error) {
	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, err
	}

	var headers []string

	if len(rows) > 0 {
		headers = rows[0]
	}

	var members []Member
	result := ImportResult{}

	now := time.Now()

	seenPhones := make(map[string]bool)

	for index, row := range rows {
		if index == 0 {
			continue
		}

		/*
			CSV esperado:

			Nome, Telefone, Status, Membro desde,
			Batizado, Data do batismo, Cargo na igreja,
			Estado civil, Congregação, Igreja de origem,
			Fez o curso de membresia, Data do curso de membresia,
			Contactado no WhatsApp, Data do contato,
			Endereço, Número, Complemento, Bairro, Cidade, Estado
		*/

		if len(row) < 20 {
			addImportError(&result, index+1, row, "estrutura da linha inválida")
			continue
		}

		baptized := parseOptionalBool(row[4])
		membershipCourseCompleted := parseOptionalBool(row[10])
		contacted := parseOptionalBool(row[12])

		memberSince := parseOptionalDate(row[3])
		baptismDate := parseOptionalDate(row[5])
		membershipCourseCompletedAt := parseOptionalDate(row[11])
		contactedAt := parseOptionalDate(row[13])

		maritalStatus := normalizeMaritalStatus(row[7])
		if !isValidMaritalStatus(maritalStatus) {
			addImportError(&result, index+1, row, "estado civil inválido: "+row[7])
			continue
		}

		congregation, validCongregation := normalizeCongregation(row[8])
		if !validCongregation {
			addImportError(&result, index+1, row, "congregação inválida: "+row[8])
			continue
		}

		member := Member{
			Name:                        row[0],
			Phone:                       row[1],
			Status:                      normalizeStatus(row[2]),
			MemberSince:                 memberSince,
			Baptized:                    baptized,
			BaptismDate:                 baptismDate,
			ChurchRole:                  row[6],
			MaritalStatus:               maritalStatus,
			OriginDenomination:          row[9],
			Congregation:                congregation,
			MembershipCourseCompleted:   membershipCourseCompleted,
			MembershipCourseCompletedAt: membershipCourseCompletedAt,
			Contacted:                   contacted,
			ContactedAt:                 contactedAt,
			Address:                     row[14],
			AddressNumber:               row[15],
			AddressComplement:           row[16],
			Neighborhood:                row[17],
			City:                        row[18],
			State:                       row[19],
			CreatedBy:                   userID,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		}

		// minimum validation
		if member.Name == "" {
			addImportError(&result, index+1, row, "nome é obrigatório")
			continue
		}

		if member.Status == "" {
			member.Status = Active
		}

		if member.Status != Active && member.Status != Inactive {
			addImportError(&result, index+1, row, "status inválido para o membro: "+member.Name)
			continue
		}

		if member.Phone != "" {
			if seenPhones[member.Phone] {
				addImportError(&result, index+1, row, "telefone duplicado no arquivo")
				continue
			}
			seenPhones[member.Phone] = true
		}

		members = append(members, member)
	}

	if result.Failed > 0 {
		jobID := uuid.NewString()

		job := ImportJob{
			ID:        jobID,
			CreatedAt: time.Now(),
			Headers:   headers,
			ErrorRows: result.Errors,
		}

		SaveImportJob(job)

		result.JobID = jobID
	}

	if len(members) == 0 {
		return result, nil
	}

	err = s.repo.CreateMany(members)
	if err != nil {
		return result, err
	}

	result.Imported = len(members)

	return result, nil
}

func (s *Service) GetByID(id int64) (Member, error) {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return Member{}, errors.New("member not found")
	}

	return member, nil
}

func parseOptionalDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	for _, layout := range []string{"2006-01-02", "02/01/2006"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

func parseOptionalBool(value string) bool {
	if value == "" {
		return false
	}

	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "sim" {
		return true
	}
	if normalized == "não" || normalized == "nao" {
		return false
	}

	parsed, err := strconv.ParseBool(normalized)
	if err != nil {
		return false
	}

	return parsed
}

func normalizeStatus(value string) Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ativo", "active":
		return Active
	case "inativo", "inactive":
		return Inactive
	default:
		return Status(value)
	}
}

func normalizeMaritalStatus(value string) MaritalStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "solteiro", "solteira", "single":
		return Single
	case "casado", "casada", "married":
		return Married
	case "divorciado", "divorciada", "divorced":
		return Divorced
	case "viúvo", "viúva", "viuvo", "viuva", "widowed":
		return Widowed
	default:
		return MaritalStatus(value)
	}
}

func normalizeCongregation(value string) (string, bool) {
	congregations := []string{
		"Sede",
		"Várzea de Cima",
		"Caraúno",
		"Cohab",
		"Quixadá",
		"Fortaleza",
		"Castelo",
		"Conjunto Esperança",
	}

	value = strings.TrimSpace(value)
	for _, congregation := range congregations {
		if strings.EqualFold(value, congregation) {
			return congregation, true
		}
	}

	return value, false
}

func addImportError(
	result *ImportResult,
	rowNumber int,
	row []string,
	message string,
) {
	result.Failed++

	result.Errors = append(
		result.Errors,
		ImportError{
			Row:   rowNumber,
			Error: message,
			Data:  row,
		},
	)
}

func isValidMaritalStatus(status MaritalStatus) bool {
	switch status {
	case Single, Married, Divorced, Widowed, Solteiro, Casado, Divorciado, Viúvo:
		return true
	default:
		return false
	}
}
