package member

import (
	"errors"
	"strings"
	"testing"
)

type serviceRepositoryStub struct {
	tx         *transactionRepositoryStub
	candidates []Member
}

func (r *serviceRepositoryStub) RunInTransaction(fn func(TransactionRepository) error) error {
	return fn(r.tx)
}

func (r *serviceRepositoryStub) FindDuplicateCandidates(_ NormalizedMember, excludeID int64) ([]Member, error) {
	return filterMembers(r.candidates, excludeID), nil
}

func (r *serviceRepositoryStub) Find(map[string]string, int, int) ([]Member, error) {
	return nil, nil
}
func (r *serviceRepositoryStub) Count(map[string]string) (int, error) { return 0, nil }
func (r *serviceRepositoryStub) GetByID(id int64) (Member, error) {
	return r.tx.GetByID(id)
}
func (r *serviceRepositoryStub) SoftDelete(int64) error { return nil }

type transactionRepositoryStub struct {
	members []Member
	created []Member
	updated []Member
	audits  []DuplicateAudit
	nextID  int64
}

func (r *transactionRepositoryStub) AcquireDuplicateLock(string) error { return nil }

func (r *transactionRepositoryStub) FindDuplicateCandidates(normalized NormalizedMember, excludeID int64) ([]Member, error) {
	result := make([]Member, 0)
	for _, candidate := range r.members {
		candidateNormalized := NormalizeMember(candidate)
		if candidate.ID != excludeID && ((normalized.Name != "" && normalized.Name == candidateNormalized.Name) ||
			(normalized.Phone != "" && normalized.Phone == candidateNormalized.Phone)) {
			result = append(result, candidate)
		}
	}
	return result, nil
}

func (r *transactionRepositoryStub) Create(member Member) (Member, error) {
	if r.nextID == 0 {
		r.nextID = 100
	}
	member.ID = r.nextID
	r.nextID++
	r.members = append(r.members, member)
	r.created = append(r.created, member)
	return member, nil
}

func (r *transactionRepositoryStub) GetByID(id int64) (Member, error) {
	for _, member := range r.members {
		if member.ID == id {
			return member, nil
		}
	}
	return Member{}, errors.New("not found")
}

func (r *transactionRepositoryStub) Update(member Member) (Member, error) {
	r.updated = append(r.updated, member)
	return member, nil
}

func (r *transactionRepositoryStub) CreateDuplicateAudits(audits []DuplicateAudit) error {
	r.audits = append(r.audits, audits...)
	return nil
}

func TestCreateReturnsConflictForHighRiskDuplicate(t *testing.T) {
	repo := newServiceRepository(Member{ID: 1, Name: "Maria", Phone: "88999999999"})
	_, err := NewService(repo).Create(Member{Name: "MARIA", Phone: "+55 88 99999-9999", CreatedBy: 9}, false)
	var conflict *DuplicateConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("Create() error = %v, want DuplicateConflictError", err)
	}
	if len(repo.tx.created) != 0 || conflict.Result.HighestRisk != RiskHigh {
		t.Fatalf("member created despite conflict: created=%d result=%+v", len(repo.tx.created), conflict.Result)
	}
}

func TestForceCreateAllowsCreationAndCreatesAudit(t *testing.T) {
	repo := newServiceRepository(Member{ID: 1, Name: "Maria", Phone: "88999999999"})
	created, err := NewService(repo).Create(Member{Name: "Maria", Phone: "88999999999", CreatedBy: 9}, true)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || len(repo.tx.created) != 1 {
		t.Fatalf("forced member was not created: %+v", created)
	}
	if len(repo.tx.audits) != 1 {
		t.Fatalf("audit count = %d, want 1", len(repo.tx.audits))
	}
	audit := repo.tx.audits[0]
	if audit.MemberID != created.ID || audit.CandidateMemberID != 1 || audit.ConfirmedBy != 9 || audit.Score != 85 {
		t.Fatalf("unexpected audit: %+v", audit)
	}
}

func TestCreateWithoutDuplicateKeepsNormalFlow(t *testing.T) {
	repo := newServiceRepository()
	created, err := NewService(repo).Create(Member{Name: "Pessoa Nova", CreatedBy: 9}, false)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || len(repo.tx.created) != 1 || len(repo.tx.audits) != 0 {
		t.Fatalf("unexpected normal create result: created=%+v audits=%v", created, repo.tx.audits)
	}
}

func TestImportComparesAgainstRowsAlreadyAcceptedFromSameFile(t *testing.T) {
	repo := newServiceRepository()
	csv := strings.Join([]string{
		"Nome,Telefone,Status,Membro desde,Batizado,Data do batismo,Cargo,Estado civil,Congregação,Origem,Curso,Data curso,Contactado,Data contato,Endereço,Número,Complemento,Bairro,Cidade,Estado",
		"Maria,88999999999,Ativo,,Não,,,Solteiro,Sede,,Não,,Não,,Rua A,10,,Centro,Quixadá,CE",
		"MARIA,+55 88 99999-9999,Ativo,,Não,,,Solteiro,Sede,,Não,,Não,,Rua A,10,,Centro,Quixadá,CE",
	}, "\n")

	result, err := NewService(repo).ImportCSV(strings.NewReader(csv), 9)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 || result.Failed != 1 || !strings.Contains(result.Errors[0].Error, "alta probabilidade") {
		t.Fatalf("unexpected import result: %+v", result)
	}
}

func newServiceRepository(members ...Member) *serviceRepositoryStub {
	for index := range members {
		if members[index].Status == "" {
			members[index].Status = Active
		}
	}
	return &serviceRepositoryStub{tx: &transactionRepositoryStub{members: members}}
}

func filterMembers(members []Member, excludeID int64) []Member {
	result := make([]Member, 0, len(members))
	for _, member := range members {
		if member.ID != excludeID {
			result = append(result, member)
		}
	}
	return result
}
