package repository

import "context"

// UnitOfWork defines the service-layer transaction or work boundary around repository access.
type UnitOfWork interface {
	// Do runs a service callback with repositories and propagates callback errors.
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

// SimpleUnitOfWork is the minimal Phase 1 work boundary without exposing the underlying connection.
type SimpleUnitOfWork struct {
	factory Factory
}

// NewUnitOfWork creates a minimal unit of work from a repository factory.
func NewUnitOfWork(factory Factory) SimpleUnitOfWork {
	return SimpleUnitOfWork{factory: factory}
}

// Do runs a service callback with repositories and propagates callback errors.
func (unit SimpleUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fn == nil {
		return NewError(ErrorCodeNotImplemented, "callback", "run unit of work", "unit of work callback is required", nil)
	}
	return fn(ctx, unit.factory.Repositories())
}
