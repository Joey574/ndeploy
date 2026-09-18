package dbu

import "github.com/Joey574/sink/v2/pkg/sink"

func SetSink(sink sink.Sink) func(dbu *Dbu) {
	return func(dbu *Dbu) {
		dbu.sink = sink
	}
}

func InheritSink(name string, s sink.Sink) func(dbu *Dbu) {
	return func(dbu *Dbu) {
		dbu.sink = sink.New(
			sink.Clone(s),
			sink.SetParent(s),
			sink.SetName(name),
		)
	}
}
