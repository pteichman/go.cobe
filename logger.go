package cobe

import (
	"github.com/cactus/go-statsd-client/statsd"
)

// Use a global statsd client.
var stats statsd.Statter

func init() {
	// err from NewNoop is always nil
	stats, _ = statsd.NewNoop()
}

// SetStatter configures a statsd client for cobe use.
func SetStatter(s statsd.Statter) {
	stats = s
}
