package sshu

import "github.com/Joey574/sink/v2/pkg/sink"

func SetSink(s sink.Sink) func(*Client) {
	return func(c *Client) {
		c.sink = s
	}
}
