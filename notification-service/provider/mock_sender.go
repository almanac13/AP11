package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type MockSender struct {
	latency time.Duration
}

func NewMockSender(latency time.Duration) *MockSender {
	return &MockSender{latency: latency}
}

func (s *MockSender) Send(to string, subject string, body string) error {
	time.Sleep(s.latency)

	if rand.Intn(10) < 3 {
		return errors.New("simulated provider temporary failure")
	}

	log.Printf("[Provider:SIMULATED] email sent to=%s subject=%q body=%q", to, subject, body)
	return nil
}
