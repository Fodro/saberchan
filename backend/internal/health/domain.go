package health

import "context"

type Service interface {
	Readiness(ctx context.Context) error
}
