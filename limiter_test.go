package goja

import (
	"testing"
)

func TestSetRateLimiter(t *testing.T) {
	t.Run("should handle when setting rate limiter to nil", func(t *testing.T) {
		r := New()
		r.SetRateLimiter(nil)
		if r.limiter == nil {
			t.Fatal("limiter should not be nil")
		}
	})
}
