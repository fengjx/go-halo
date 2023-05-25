package worker

import "time"

type Option func(*Pool)

func WithCapacity(capacity int) Option {
	return func(p *Pool) {
		p.capacity = capacity
	}
}

func WithSubmitTimeout(timeout time.Duration) Option {
	return func(p *Pool) {
		p.submitTimeout = timeout
	}
}
