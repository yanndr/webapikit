package endpoint

import "context"

type Endpoint[T, K any] func(ctx context.Context, request T) (response K, err error)
