package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/yanndr/webapikit/endpoint"
	"io"
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
		enc:    EncodeResponse[K],
		logger: l,
	}
}

// ServeHTTP implements http.Handler.
func (s *Server[T, K]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := s.dec(ctx, r)
	if err != nil {
		s.logger.Printf("err server: %w", err)
		DefaultErrorEncoder(ctx, err, w)
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

func EncodeResponse[T any](_ context.Context, w http.ResponseWriter, response T) error {
	return json.NewEncoder(w).Encode(response)
}

type EmptyRequest struct {
}

func DecodeEmptyRequest(_ context.Context, r *http.Request) (*EmptyRequest, error) {
	return &EmptyRequest{}, nil
}

func DecodeRequest[T any](_ context.Context, r *http.Request) (*T, error) {
	var request T
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return &request, nil
}

func EncodeRequest[T any](_ context.Context, r *http.Request, request *T) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = io.NopCloser(&buf)
	return nil
}
