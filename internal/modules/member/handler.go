package member

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Robi9/church-manager/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service MemberService
}

type MemberService interface {
	Create(Member, bool) (Member, error)
	CheckDuplicates(Member, int64) (DuplicateCheckResult, error)
	Find(map[string]string, int, int) ([]Member, PaginationMeta, error)
	Update(int64, Member, bool, int64) (Member, error)
	SoftDelete(int64) error
	ImportCSV(io.Reader, int64) (ImportResult, error)
	ConfirmImportDuplicate(string, int, int64) (Member, error)
	DismissImportDuplicate(string, int) error
	GetByID(int64) (Member, error)
}

func NewHandler(service MemberService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var request MemberMutationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	userID, err := authenticatedUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	request.Member.CreatedBy = userID

	result, err := h.service.Create(request.Member, request.ForceCreate)
	if err != nil {
		handleMutationError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

func (h *Handler) CheckDuplicates(c *gin.Context) {
	var request DuplicateCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.CheckDuplicates(request.Member, request.ExcludeMemberID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	response.Success(c, http.StatusOK, result)
}

func (h *Handler) Find(c *gin.Context) {
	name := c.Query("name")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filters := map[string]string{
		"name":   name,
		"status": status,
	}

	data, meta, err := h.service.Find(filters, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"data": data,
		"meta": meta,
	})
}

func (h *Handler) Update(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}

	var request MemberMutationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	userID, err := authenticatedUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	result, err := h.service.Update(id, request.Member, request.ForceCreate, userID)
	if err != nil {
		handleMutationError(c, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *Handler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}

	err = h.service.SoftDelete(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "member deactivated successfully",
	})
}

func (h *Handler) Import(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("file is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	userID, err := authenticatedUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	result, err := h.service.ImportCSV(file, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

type confirmImportDuplicateRequest struct {
	Row int `json:"row"`
}

func (h *Handler) ConfirmImportDuplicate(c *gin.Context) {
	var request confirmImportDuplicateRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Row <= 1 {
		response.Error(c, http.StatusBadRequest, errors.New("valid import row is required"))
		return
	}
	userID, err := authenticatedUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}
	created, err := h.service.ConfirmImportDuplicate(c.Param("jobID"), request.Row, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	response.Success(c, http.StatusCreated, created)
}

func (h *Handler) DismissImportDuplicate(c *gin.Context) {
	var request confirmImportDuplicateRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Row <= 1 {
		response.Error(c, http.StatusBadRequest, errors.New("valid import row is required"))
		return
	}
	if err := h.service.DismissImportDuplicate(c.Param("jobID"), request.Row); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "import row dismissed"})
}

func authenticatedUserID(c *gin.Context) (int64, error) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user not authenticated")
	}
	switch userID := value.(type) {
	case float64:
		return int64(userID), nil
	case int64:
		return userID, nil
	case int:
		return int64(userID), nil
	default:
		return 0, errors.New("invalid authenticated user")
	}
}

func handleMutationError(c *gin.Context, err error) {
	var conflict *DuplicateConflictError
	if errors.As(err, &conflict) {
		c.JSON(http.StatusConflict, response.Response{
			Data:  conflict.Result,
			Error: conflict.Error(),
		})
		return
	}
	if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidStatus) {
		response.Error(c, http.StatusBadRequest, err)
		return
	}
	if err.Error() == "member not found" {
		response.Error(c, http.StatusNotFound, err)
		return
	}
	response.Error(c, http.StatusInternalServerError, err)
}

func (h *Handler) DownloadTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header(
		"Content-Disposition",
		`attachment; filename="modelo_membros.csv"`,
	)

	csvContent := `Nome,Telefone,Status,Membro desde,Batizado,Data do batismo,Cargo na igreja,Estado civil,Congregação,Igreja de origem,Fez o curso de membresia,Data do curso de membresia,Contactado no WhatsApp,Data do contato,Endereço,Número,Complemento,Bairro,Cidade,Estado
Maria,88999999999,Ativo,2024-01-01,Sim,2024-02-01,Líder,Solteiro,Sede,Batista,Sim,2024-02-15,Sim,2024-03-01,Rua das Flores,123,Apto 1,Centro,Fortaleza,CE
`

	c.String(http.StatusOK, csvContent)
}

func (h *Handler) DownloadImportErrors(c *gin.Context) {
	jobID := c.Param("jobID")

	job, ok := GetImportJob(jobID)
	if !ok {
		c.JSON(404, gin.H{
			"error": "import job not found",
		})
		return
	}

	c.Header(
		"Content-Type",
		"text/csv",
	)

	c.Header(
		"Content-Disposition",
		"attachment; filename=import_errors.csv",
	)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := append(job.Headers, "erro")

	err := writer.Write(headers)
	if err != nil {
		return
	}

	for _, errRow := range job.ErrorRows {
		row := normalizeRow(
			errRow.Data,
			len(job.Headers),
		)

		row = append(
			row,
			errRow.Error,
		)

		err := writer.Write(row)
		if err != nil {
			return
		}
	}
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	member, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, errors.New("member not found"))
		return
	}

	response.Success(c, http.StatusOK, member)
}
