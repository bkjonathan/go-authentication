package utils

import (
	"encoding/json"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Normalizer is a request that tidies itself up before it is judged.
type Normalizer interface {
	Normalize()
}

// BindJSON decodes the body, normalises it, and only then validates. The order
// matters: " Ada@Example.com " is an address a real keyboard produces, and the
// trimming that makes it valid has to happen before `email` gets to reject it.
func BindJSON(ctx *gin.Context, request any) *apperror.Error {
	if err := json.NewDecoder(ctx.Request.Body).Decode(request); err != nil {
		return apperror.BadRequest("invalid_request", "The request body could not be read.").Wrap(err)
	}
	if normalizer, ok := request.(Normalizer); ok {
		normalizer.Normalize()
	}
	if err := binding.Validator.ValidateStruct(request); err != nil {
		return BindingError(err)
	}
	return nil
}
