package response

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalErrorsNeverExposeDetails(t *testing.T) {
	for _, test := range []struct {
		status int
		err    error
	}{
		{500, errors.New("password=secret member Maria")},
		{400, &pq.Error{Message: "duplicate member Maria", Detail: "phone=123"}},
	} {
		res := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(res)
		ctx.Request = httptest.NewRequest("POST", "/members", nil)
		Error(ctx, test.status, test.err)
		body := res.Body.String()
		if strings.Contains(body, "Maria") || strings.Contains(body, "secret") || strings.Contains(body, "123") {
			t.Fatalf("sensitive error exposed: %s", body)
		}
	}
}
