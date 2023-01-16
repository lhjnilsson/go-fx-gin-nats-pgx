package marketdata

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/instrument"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewMarketDataStreamer(lc fx.Lifecycle, instrumentStreams *instrument.InstrumentStreams, md marketdata.Client, ohlcStorage OHLCStorage, log *zap.Logger) error {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	js.AddConsumer("instrument", &nats.ConsumerConfig{
		Durable:        "marketdata",
		DeliverSubject: "marketdata",
		AckPolicy:      nats.AckExplicitPolicy,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting market data streamer")
			_, err := js.Subscribe(instrumentStreams.InstrumentAdded, func(msg *nats.Msg) {
				symbol := string(msg.Data)
				log.Info("Received instrument added", zap.String("instrument", symbol))

				bar, err := md.GetLatestBar(symbol)
				if err != nil {
					log.Error("Failed to get bar", zap.Error(err))
					return
				}
				ohlc := OHLC{
					Open:      bar.Open,
					High:      bar.High,
					Low:       bar.Low,
					Close:     bar.Close,
					Volume:    bar.Volume,
					Timestamp: bar.Timestamp,
				}
				err = ohlcStorage.Add(context.TODO(), symbol, &ohlc)
				if err != nil {
					log.Error("Failed to add OHLC", zap.Error(err))
					return
				}
				log.Info("Added OHLC", zap.String("Symbol", symbol), zap.Any("OHLC", ohlc))
				err = msg.Ack()
				if err != nil {
					log.Error("Failed to ack message", zap.Error(err))
				}
			})
			if err != nil {
				return err
			}
			_, err = js.Subscribe(instrumentStreams.InstrumentRemoved, func(msg *nats.Msg) {
				log.Info("Received instrument removed", zap.String("instrument", string(msg.Data)))
				symbol := string(msg.Data)
				err = ohlcStorage.Remove(context.TODO(), symbol)
				if err != nil {
					log.Error("Failed to remove OHLC", zap.Error(err))
					return
				}
				log.Info("Removed OHLC", zap.String("Symbol", symbol))
				err = msg.Ack()
				if err != nil {
					log.Error("Failed to ack message", zap.Error(err))
				}
			})
			if err != nil {
				return err
			}
			log.Info("Market data streamer started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			nc.Close()
			return nil
		},
	})
	return nil
}
