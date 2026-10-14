package utils

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
)

// GetDisplayErrorMessage returns the display error message for given error
func GetDisplayErrorMessage(err *errs.Error) string {
	if err.Code() == errs.ErrIncompleteOrIncorrectSubmission.Code() && len(err.BaseError) > 0 {
		var validationErrors validator.ValidationErrors
		ok := errors.As(err.BaseError[0], &validationErrors)

		if ok {
			for _, err := range validationErrors {
				return getValidationErrorText(err)
			}
		}
	}

	return err.Error()
}

// GetJsonErrorResult returns error response in json format
func GetJsonErrorResult(err *errs.Error, path string) map[string]any {
	return core.O{
		"success":      false,
		"errorCode":    err.Code(),
		"errorMessage": GetDisplayErrorMessage(err),
		"path":         path,
	}
}

// PrintJsonSuccessResult writes success response in json format to current http context
func PrintJsonSuccessResult(c *core.WebContext, result any) {
	c.JSON(http.StatusOK, core.O{
		"success": true,
		"result":  result,
	})
}

// PrintDataSuccessResult writes success response in custom content type to current http context
func PrintDataSuccessResult(c *core.WebContext, contentType string, fileName string, result []byte) {
	if fileName != "" {
		c.Header("Content-Disposition", "attachment;filename="+fileName)
	}

	c.Data(http.StatusOK, contentType, result)
}

// PrintJsonErrorResult writes error response in json format to current http context
func PrintJsonErrorResult(c *core.WebContext, err *errs.Error) {
	c.SetResponseError(err)

	result := GetJsonErrorResult(err, c.Request.URL.Path)

	if err.Context != nil {
		result["context"] = err.Context
	}

	c.AbortWithStatusJSON(err.HttpStatusCode, result)
}

// PrintJSONRPCSuccessResult writes success response in JSON-RPC format to current http context
func PrintJSONRPCSuccessResult(c *core.WebContext, jsonRPCRequest *core.JSONRPCRequest, result any) {
	c.JSON(http.StatusOK, core.NewJSONRPCResponse(jsonRPCRequest.ID, result))
}

// PrintJSONRPCErrorResult writes error response in JSON-RPC format to current http context
func PrintJSONRPCErrorResult(c *core.WebContext, jsonRPCRequest *core.JSONRPCRequest, err *errs.Error) {
	c.SetResponseError(err)

	var id any

	if jsonRPCRequest != nil {
		id = jsonRPCRequest.ID
	}

	jsonRPCError := core.JSONRPCInternalError

	if err.Code() == errs.ErrIncompleteOrIncorrectSubmission.Code() {
		jsonRPCError = core.JSONRPCParseError
	} else if err.Code() == errs.ErrApiNotFound.Code() {
		jsonRPCError = core.JSONRPCMethodNotFoundError
	} else if err.Code() == errs.ErrParameterInvalid.Code() {
		jsonRPCError = core.JSONRPCInvalidParamsError
	}

	c.AbortWithStatusJSON(err.HttpStatusCode, core.NewJSONRPCErrorResponseWithCause(id, jsonRPCError, GetDisplayErrorMessage(err)))
}

// PrintDataErrorResult writes error response in custom content type to current http context
func PrintDataErrorResult(c *core.WebContext, contentType string, err *errs.Error) {
	c.SetResponseError(err)
	c.Data(err.HttpStatusCode, contentType, []byte(GetDisplayErrorMessage(err)))
	c.Abort()
}

// SetEventStreamHeader sets the headers for event stream response
func SetEventStreamHeader(c *core.WebContext) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
}

func WriteEventStreamJsonSuccessResult(c *core.WebContext, result any) {
	data, err := json.Marshal(result)

	if err != nil {
		c.Abort()
		return
	}

	_, err = c.Writer.WriteString("data: " + string(data) + "\n\n")

	if err != nil {
		c.Abort()
		return
	}

	c.Writer.Flush()
}

func WriteEventStreamJsonErrorResult(c *core.WebContext, originalErr *errs.Error) {
	c.SetResponseError(originalErr)

	result := GetJsonErrorResult(originalErr, c.Request.URL.Path)

	if originalErr.Context != nil {
		result["context"] = originalErr.Context
	}

	data, err := json.Marshal(result)

	if err != nil {
		c.Abort()
		return
	}

	_, err = c.Writer.WriteString("data: " + string(data) + "\n\n")

	if err != nil {
		c.Abort()
		return
	}

	c.Writer.Flush()
}

// WriteEventStreamJSON is the checked variant for request-owned streams. The
// caller owns cancellation and serializes it with comments and terminal frames.
func WriteEventStreamJSON(c *core.WebContext, result any) error {
	data, err := jsonv1.Marshal(result)
	if err != nil {
		return err
	}
	return writeEventStreamFrame(c, "data: "+string(data)+"\n\n")
}

func WriteEventStreamComment(c *core.WebContext, comment string) error {
	return writeEventStreamFrame(c, ": "+comment+"\n\n")
}

func writeEventStreamFrame(c *core.WebContext, frame string) error {
	n, err := c.Writer.WriteString(frame)
	if err == nil && n != len(frame) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return err
	}
	// Gin's Flush discards the underlying FlushError. Unwrap its bookkeeping
	// wrapper after WriteString so a disconnected transport remains observable.
	writer := http.ResponseWriter(c.Writer)
	if wrapped, ok := writer.(interface{ Unwrap() http.ResponseWriter }); ok {
		writer = wrapped.Unwrap()
	}
	return http.NewResponseController(writer).Flush()
}

func getValidationErrorText(err validator.FieldError) string {
	fieldName := GetFirstLowerCharString(err.Field())

	switch err.Tag() {
	case "required":
		return errs.GetParameterIsRequiredMessage(fieldName)
	case "max":
		if isIntegerParameter(err.Kind()) {
			return errs.GetParameterMustLessThanMessage(fieldName, err.Param())
		} else if isStringParameter(err.Kind()) {
			return errs.GetParameterMustLessThanCharsMessage(fieldName, err.Param())
		}
	case "min":
		if isIntegerParameter(err.Kind()) {
			return errs.GetParameterMustMoreThanMessage(fieldName, err.Param())
		} else if isStringParameter(err.Kind()) {
			return errs.GetParameterMustMoreThanCharsMessage(fieldName, err.Param())
		}
	case "len":
		return errs.GetParameterLengthNotEqualMessage(fieldName, err.Param())
	case "notBlank":
		return errs.GetParameterNotBeBlankMessage(fieldName)
	case "validUsername":
		return errs.GetParameterInvalidUsernameMessage(fieldName)
	case "validEmail":
		return errs.GetParameterInvalidEmailMessage(fieldName)
	case "validCurrency":
		return errs.GetParameterInvalidCurrencyMessage(fieldName)
	case "validHexRGBColor":
		return errs.GetParameterInvalidHexRGBColorMessage(fieldName)
	case "validAmountFilter":
		return errs.GetParameterInvalidAmountFilterMessage(fieldName)
	case "validTransactionAmount":
		return errs.GetParameterInvalidTransactionAmountMessage(fieldName)
	case "validTagFilter":
		return errs.GetParameterInvalidTagFilterMessage(fieldName)
	}

	return errs.GetParameterInvalidMessage(fieldName)
}

func isIntegerParameter(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isStringParameter(kind reflect.Kind) bool {
	return kind == reflect.String
}
