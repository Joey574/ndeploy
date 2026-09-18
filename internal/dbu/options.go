package dbu

import "github.com/Joey574/sink/v2/pkg/sink"

func SetSink(sink sink.Sink) func(dbu *Dbu) {
	return func(dbu *Dbu) {
		dbu.sink = sink
	}
}
