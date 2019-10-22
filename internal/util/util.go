package util

import (
	"log"
)

// OptionalDebug returns a function that will noop when debugEnabled=false, and
// will log the given msg using `log.Print` or `log.Printf` (depending on the
// number arguments given) debugEnabled=true
func OptionalDebug(debugEnabled bool) func(string, ...interface{}) {
	if !debugEnabled {
		// return a noop function
		return func(msg string, args ...interface{}) {
			return
		}
	}
	// return the real debug function
	return func(msg string, args ...interface{}) {
		if args == nil {
			log.Print(msg)
		}
		log.Printf(msg, args...)
	}
}
