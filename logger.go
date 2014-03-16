package cobe

import (
	"github.com/cactus/go-statsd-client/statsd"
	logging "github.com/op/go-logging"
)

// Call this clog instead of log so it doesn't confuse goimports. I'll
// rename this if it works better in the future.
var clog = logging.MustGetLogger("cobe")

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
