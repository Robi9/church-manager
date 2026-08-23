package member

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type handlerServiceStub struct {
	createResult Member
	createError  error
	forceCreate  bool
}

func (s *handlerServiceStub) Create(_ Member, force bool) (Member, error) {
	s.forceCreate = force
	return s.createResult, s.createError
}
func (s *handlerServiceStub) CheckDuplicates(Member, int64) (DuplicateCheckResult, error) {
	return DuplicateCheckResult{}, nil
}
func (s *handlerServiceStub) Find(map[string]string, int, int) ([]Member, PaginationMeta, error) {
	return nil, PaginationMeta{}, nil
}
func (s *handlerServiceStub) Update(int64, Member, bool, int64) (Member, error) {
	return Member{}, nil
}
func (s *handlerServiceStub) SoftDelete(int64) error { return nil }
func (s *handlerServiceStub) ImportCSV(io.Reader, int64) (ImportResult, error) {
	return ImportResult{}, nil
}
func (s *handlerServiceStub) GetByID(int64) (Member, error) { return Member{}, nil }

func TestCreateHandlerReturns409WithDuplicateCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	duplicate := DuplicateCheckResult{
		HasPossibleDuplicates: true,
		HighestRisk:           RiskHigh,
		Candidates: []DuplicateCandidate{{
			MemberID: 1, Name: "Maria", Phone: "*******9999", Score: 85, Risk: RiskHigh,
		}},
	}
	service := &handlerServiceStub{createError: &DuplicateConflictError{Result: duplicate}}
	recorder := performCreateRequest(t, service, `{"name":"Maria","phone":"88999999999"}`)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data DuplicateCheckResult `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.HighestRisk != RiskHigh || len(body.Data.Candidates) != 1 {
		t.Fatalf("unexpected conflict response: %+v", body.Data)
	}
}

func TestCreateHandlerPassesForceCreateConfirmation(t *testing.T) {
	service := &handlerServiceStub{createResult: Member{ID: 10, Name: "Maria"}}
	recorder := performCreateRequest(t, service, `{"name":"Maria","force_create":true}`)
	if recorder.Code != http.StatusCreated || !service.forceCreate {
		t.Fatalf("status=%d force=%v body=%s", recorder.Code, service.forceCreate, recorder.Body.String())
	}
}

func performCreateRequest(t *testing.T, service *handlerServiceStub, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.POST("/members", func(c *gin.Context) {
		c.Set("user_id", float64(9))
		NewHandler(service).Create(c)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/members", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}
