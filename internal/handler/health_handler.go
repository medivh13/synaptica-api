package handler

import (
	"encoding/json"
	"log"
	stdhttp "net/http"
)

type HealthHandler struct {
	serviceName string
	responder   Responder
}

type Responder interface {
	WriteSuccess(w stdhttp.ResponseWriter, status int, data any)
	WriteError(w stdhttp.ResponseWriter, status int, message string)
}

func NewHealthHandler(serviceName string) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		responder:   defaultResponder{},
	}
}

func (h *HealthHandler) WithResponder(responder Responder) *HealthHandler {
	h.responder = responder
	return h
}

func (h *HealthHandler) Check(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	h.responder.WriteSuccess(w, stdhttp.StatusOK, map[string]string{
		"status":  "ok",
		"service": h.serviceName,
	})
}

type defaultResponder struct{}

func (defaultResponder) WriteSuccess(w stdhttp.ResponseWriter, status int, data any) {
	writeJSON(w, status, successResponse{Data: data})
}

func (defaultResponder) WriteError(w stdhttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Message: message}})
}

type successResponse struct {
	Data any `json:"data"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message string `json:"message"`
}

func writeJSON(w stdhttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}
