package cobe

import (
	logging "github.com/op/go-logging"
)

// Call this clog instead of log so it doesn't confuse goimports. I'll
// rename this if it works better in the future.
//
// This logger is used only within the core library. Commands
// define their own loggers.
//
var clog = logging.MustGetLogger("cobe")
