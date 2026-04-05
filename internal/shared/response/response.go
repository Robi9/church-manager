package response

import "github.com/gin-gonic/gin"

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
	c.JSON(status, Response{
		Data:  nil,
		Error: err.Error(),
	})
}
