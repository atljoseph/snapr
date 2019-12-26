package util

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// NewErrGroup is a structured way of initializing an errogroup
func NewErrGroup() (*errgroup.Group, context.Context) {
	// get a new err group to wait on goroutine group completion
	// and catch errors
	// different than wait groups!
	ctx := context.Background()
	// ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return errgroup.WithContext(ctx)
}
