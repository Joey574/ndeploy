package engine

import "github.com/Joey574/sink/v2/pkg/sink"

func SetSink(s sink.Sink) func(*Engine) {
	return func(e *Engine) {
		e.sink = s
	}
}
