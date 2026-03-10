package messages

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    string `json:"error"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, err Error) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(err)
}

func BadRequest(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusBadRequest, Error{
		Code:    code,
		Message: message,
	})
}

func Unauthorized(w http.ResponseWriter, code, message string) {
	WriteError(w, http.StatusUnauthorized, Error{
		Code:    code,
		Message: message,
	})
}

func InternalError(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, Error{
		Code:    "internal_error",
		Message: "internal server error",
	})
}