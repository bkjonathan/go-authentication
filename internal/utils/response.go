package utils

import (
	"net/http"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorEnvelope `json:"error,omitempty"`
}

type ErrorEnvelope struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func OK(c *gin.Context, data any) {
	Respond(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	Respond(c, http.StatusCreated, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Respond(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Success: true, Data: data})
}

// Fail renders a failure in the same envelope a success uses, so a client has
// one shape to parse rather than two.
func Fail(c *gin.Context, err *apperror.Error) {
	c.AbortWithStatusJSON(err.Status, Envelope{
		Success: false,
		Error: &ErrorEnvelope{
			Code:    err.Code,
			Message: err.Message,
			Fields:  err.Fields,
		},
	})
}
