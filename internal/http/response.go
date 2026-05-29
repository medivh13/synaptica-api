package http

import (
	"encoding/json"
	"log"
	stdhttp "net/http"
)

type successResponse struct {
	Data any `json:"data"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message string `json:"message"`
}

type Responder struct{}

func (Responder) WriteSuccess(w stdhttp.ResponseWriter, status int, data any) {
	WriteSuccess(w, status, data)
}

func (Responder) WriteError(w stdhttp.ResponseWriter, status int, message string) {
	WriteError(w, status, message)
}

func WriteJSON(w stdhttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func WriteSuccess(w stdhttp.ResponseWriter, status int, data any) {
	WriteJSON(w, status, successResponse{Data: data})
}

func WriteError(w stdhttp.ResponseWriter, status int, message string) {
	WriteJSON(w, status, errorResponse{Error: errorBody{Message: message}})
}
