package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	headerContentTypeKey  = "Content-Type"
	headerContentTypeJSON = "application/json"
)

type jsonError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details"`
}

func HTTPErrorHandler(w http.ResponseWriter, err error) error {
	var resultErr *Error
	if !errors.As(err, &resultErr) {
		writeErr := writeJSONError(w, jsonError{
			Message: msgInternalServerError,
			Code:    http.StatusInternalServerError,
			Details: "",
		})
		if writeErr != nil {
			return fmt.Errorf("write json error: %w", writeErr)
		}

		return nil
	}

	var jsonErr jsonError
	switch resultErr.Status() {
	case ErrorStatusNotFound:
		jsonErr = jsonError{
			Message: resultErr.message,
			Code:    http.StatusNotFound,
			Details: resultErr.Error(),
		}
	case ErrorStatusUnauthenticated:
		jsonErr = jsonError{
			Message: resultErr.message,
			Code:    http.StatusUnauthorized,
			Details: resultErr.Error(),
		}
	case ErrorStatusInvalidArgument:
		jsonErr = jsonError{
			Message: resultErr.message,
			Code:    http.StatusBadRequest,
			Details: resultErr.Error(),
		}
	case ErrorStatusForbidden:
		jsonErr = jsonError{
			Message: resultErr.message,
			Code:    http.StatusForbidden,
			Details: resultErr.Error(),
		}
	default:
		jsonErr = jsonError{
			Message: msgInternalServerError,
			Code:    http.StatusInternalServerError,
			Details: "",
		}
	}

	err = writeJSONError(w, jsonErr)
	if err != nil {
		return fmt.Errorf("write json error: %w", err)
	}

	return nil
}

func writeJSONError(w http.ResponseWriter, jsonErr jsonError) error {
	w.Header().Set(headerContentTypeKey, headerContentTypeJSON)
	w.WriteHeader(jsonErr.Code)

	err := json.NewEncoder(w).Encode(jsonErr)
	if err != nil {
		return fmt.Errorf("encode to json error: %w", err)
	}
	return nil
}
