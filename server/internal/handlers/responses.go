package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type ErrorJSON struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func errorHandler(w http.ResponseWriter, err error) {
	message := err.Error()
	statusCode := http.StatusInternalServerError

	if errors.Is(err, gorm.ErrRecordNotFound) {
		message = http.StatusText(http.StatusNotFound)
		statusCode = http.StatusNotFound
	}

	e := ErrorJSON{
		Message:    message,
		StatusCode: statusCode,
	}

	data, err := json.Marshal(e)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
	}

	http.Error(w, string(data), statusCode)
}

func responseHandler(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(data); err != nil {
		errorHandler(w, err)
		return
	}
}
