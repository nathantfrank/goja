package goja

import (
	"golang.org/x/time/rate"
)

// SetRateLimiter sets the rate limiter
func (self *Runtime) SetRateLimiter(limiter *rate.Limiter) {
	if limiter == nil {
		limiter = rate.NewLimiter(rate.Inf, maxInt)
	}
	self.limiter = limiter
	self.fillBucket()
}

const burstDivisor = 5

func (self *Runtime) fillBucket() {
	self.limiterTicksLeft = self.limiter.Burst() / burstDivisor
}
