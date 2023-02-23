package http

import (
	"context"
	"encoding/json"
	"github.com/yanndr/webapikit/endpoint"
	"log"
	"net/http"
)

type Decoder[T any] func(ctx context.Context, r *http.Request) (*T, error)
type Encoder[T any] func(ctx context.Context, w http.ResponseWriter, response T) error

type Server[T, K any] struct {
	e      endpoint.Endpoint[*T, K]
	dec    Decoder[T]
	enc    Encoder[K]
	logger *log.Logger
}

func NewServer[T, K any](e endpoint.Endpoint[*T, K], decoder Decoder[T], l *log.Logger) *Server[T, K] {
	return &Server[T, K]{
		e:      e,
		dec:    decoder,
		enc:    encodeResponse[K],
		logger: l,
	}
}

// ServeHTTP implements http.Handler.
func (s *Server[T, K]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := s.dec(ctx, r)
	if err != nil {
		s.logger.Printf("err server: %w", err)
		return
	}

	response, err := s.e(ctx, request)
	if err != nil {
		DefaultErrorEncoder(ctx, err, w)
		return
	}

	if err := s.enc(ctx, w, response); err != nil {
		DefaultErrorEncoder(ctx, err, w)
		return
	}
}

func DefaultErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	contentType, body := "text/plain; charset=utf-8", []byte(err.Error())
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			contentType, body = "application/json; charset=utf-8", jsonBody
		}
	}
	w.Header().Set("Content-Type", contentType)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(body)
}

func encodeResponse[T any](_ context.Context, w http.ResponseWriter, response T) error {
	return json.NewEncoder(w).Encode(response)
}
