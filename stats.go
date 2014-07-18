package cobe

// Use a global statter, by default noop.
var stats Statter = &NoopStatter{}

// A bare minimum interface similar to Statter interface
// in github.com/cactus/go-statsd-client.
type Statter interface {
	Inc(stat string, value int64, rate float32) error
	Timing(stat string, delta int64, rate float32) error
}

// A noop statter.
type NoopStatter struct{}

// Increments a counter.
func (s *NoopStatter) Inc(stat string, value int64, rate float32) error {
	return nil
}

// Submits timing.
func (s *NoopStatter) Timing(stat string, delta int64, rate float32) error {
	return nil
}

// Allows to enable a user provided global statter.
// Statters must implement the Statter interface such as
// in github.com/cactus/go-statsd-client.
func SetStatter(s Statter) {
	stats = s
}
