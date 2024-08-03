//go:build js
// +build js

package flock

import (
	"context"
	"time"
)

type Flock struct{}

func New(_ string) *Flock {
	return &Flock{}
}

func (f *Flock) Close() error {
	return nil
}

func (f *Flock) Lock() error {
	return nil
}

func (f *Flock) Locked() bool {
	return false
}

func (f *Flock) Path() string {
	return ""
}

func (f *Flock) RLock() error {
	return nil
}

func (f *Flock) RLocked() bool {
	return false
}

func (f *Flock) String() string {
	return ""
}

func (f *Flock) TryLock() (bool, error) {
	return false, nil
}

func (f *Flock) TryLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return false, nil
}

func (f *Flock) TryRLock() (bool, error) {
	return false, nil
}

func (f *Flock) TryRLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return false, nil
}

func (f *Flock) Unlock() error {
	return nil
}
