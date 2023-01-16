package instrument

import "github.com/nats-io/nats.go"

type InstrumentStreams struct {
	InstrumentAdded   string
	InstrumentRemoved string
}

func NewInstrumentStream() (*InstrumentStreams, error) {
	streams := &InstrumentStreams{
		InstrumentAdded:   "instrument.added",
		InstrumentRemoved: "instrument.removed",
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	js.AddStream(&nats.StreamConfig{
		Name:     "instrument",
		Subjects: []string{streams.InstrumentAdded, streams.InstrumentRemoved},
	})
	return streams, nil
}
