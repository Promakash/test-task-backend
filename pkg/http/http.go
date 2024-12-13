package http

import (
	"encoding/json"
	"io"
	"net/http"
)

type Handler func(*http.Request) Response

type Response interface {
	StatusCode() int
	GetPayload() any
}

type BasicResponse struct {
	Payload    any
	statusCode int
}

func (r BasicResponse) StatusCode() int {
	return r.statusCode
}

func (r BasicResponse) GetPayload() any {
	return r.Payload
}

func OK(payload any) *BasicResponse {
	return &BasicResponse{
		statusCode: http.StatusOK,
		Payload:    payload,
	}
}

type ErrorResponse struct {
	Message    string `json:"message"`
	err        error
	statusCode int
}

func (r ErrorResponse) StatusCode() int {
	return r.statusCode
}

func (r ErrorResponse) GetPayload() any {
	return r
}

func BadRequest(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusBadRequest,
		Message:    err.Error(),
		err:        err,
	}
}

func NotFound(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusNotFound,
		Message:    err.Error(),
		err:        err,
	}
}

func AddHandler(
	mountMethod func(pattern string, h http.HandlerFunc),
	pattern string,
	handler Handler,
) {
	mountMethod(pattern, converter(handler))
}

func converter(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(r)
		if resp == nil {
			return
		}

		writeResponse(w, resp)
	}
}

func writeResponse(w http.ResponseWriter, response Response) {
	r, err := json.Marshal(response.GetPayload())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(response.StatusCode())

	_, _ = w.Write(r)
}

func ExtractPayload[T any](resp *http.Response) (T, error) {
	var payload T

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	if resp.StatusCode != http.StatusOK {
		var errRes ErrorResponse
		if err := json.Unmarshal(body, &errRes); err != nil {
			return payload, err
		}
		return payload, errRes.err
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return payload, err
	}

	return payload, nil
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}