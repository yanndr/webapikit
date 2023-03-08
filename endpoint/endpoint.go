package endpoint

import "context"

type Endpoint[T, K any] func(ctx context.Context, request T) (response K, err error)

type Middleware func(Endpoint[interface{}, interface{}]) Endpoint[interface{}, interface{}]
