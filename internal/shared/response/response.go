package response

import (
	"errors"
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Data:  data,
		Error: nil,
	})
}

func Error(c *gin.Context, status int, err error) {
	message := err.Error()
	var dbErr *pq.Error
	safeMessages := map[string]bool{
		"name is required": true, "invalid status": true,
		"invalid credentials": true, "email and password are required": true,
		"password must be at least 12 characters": true, "email already registered": true,
		"member not found": true, "member already inactive": true,
		"user not authenticated": true, "invalid member id": true, "invalid id": true,
		"file is required": true, "file must be at most 5 MB": true,
		"valid import row is required":                    true,
		"import row not found or already being processed": true,
		"import row is not eligible for confirmation":     true,
		"import row is not eligible for dismissal":        true,
	}
	if status >= 500 || errors.As(err, &dbErr) || !safeMessages[message] {
		// Do not log the error text: database errors can contain submitted personal data.
		log.Printf("request failed: method=%s route=%s error_type=%v", c.Request.Method, c.FullPath(), reflect.TypeOf(err))
		message = "Não foi possível concluir a operação. Tente novamente."
	}
	c.JSON(status, Response{Data: nil, Error: message})
}
