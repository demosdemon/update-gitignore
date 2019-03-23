package async

import "context"

type Task = func(ctx context.Context) (err error)
