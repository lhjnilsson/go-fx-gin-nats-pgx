package instrument

import (
	"github.com/alpacahq/alpaca-trade-api-go/v2/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func RegisterRoutes(routeGroup *gin.RouterGroup, log *zap.Logger, md marketdata.Client, trade alpaca.Client, instrumentStorage InstrumentStorage, stream nats.JetStream) {
	r := &Routes{log: log, marketdata: md, trade: trade, instrumentStorage: instrumentStorage, stream: stream}
	routeGroup.GET("/:symbol", r.GetInstrument)
	routeGroup.POST("/:symbol", r.AddInstrument)
	routeGroup.DELETE("/:symbol", r.RemoveInstrument)
}

type Routes struct {
	log               *zap.Logger
	marketdata        marketdata.Client
	trade             alpaca.Client
	instrumentStorage InstrumentStorage
	stream            nats.JetStream
}

// GetInstrument godoc
//
//	@Summary	Get an Instrument
//	@Tags		instrument
//	@Accept		json
//	@Produce	json
//	@Param		symbol	path		string	true	"Instrument Symbol"
//	@Success	200		{object}	Instrument
//	@Router		/instruments/{symbol} [get]
func (r *Routes) GetInstrument(c *gin.Context) {
	instrument, err := r.instrumentStorage.Get(c, c.Param("symbol"))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, instrument)
}

// AddInstrument godoc
//
//	@Summary	Add an Instrument
//	@Tags		instrument
//	@Accept		json
//	@Produce	json
//	@Param		symbol	path		string	true	"Instrument Symbol"
//	@Success	200		{object}	Instrument
//	@Router		/instruments/{symbol} [post]
func (r *Routes) AddInstrument(c *gin.Context) {
	tradeInstrument, err := r.trade.GetAsset(c.Param("symbol"))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	instrument := Instrument{
		ID:       tradeInstrument.ID,
		Symbol:   tradeInstrument.Symbol,
		Name:     tradeInstrument.Name,
		Class:    tradeInstrument.Class,
		Exchange: tradeInstrument.Exchange,
	}
	err = r.instrumentStorage.Add(c, &instrument)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	r.log.Info("Added instrument", zap.String("symbol", instrument.Symbol))
	r.stream.Publish("instrument.added", []byte(instrument.Symbol))
	c.JSON(200, instrument)
}

// RemoveInstrument godoc
//
//	@Summary	Remove an Instrument
//	@Tags		instrument
//	@Produce	json
//	@Param		symbol	path	string	true	"Instrument Symbol"
//	@Router		/instruments/{symbol} [delete]
func (r *Routes) RemoveInstrument(c *gin.Context) {
	err := r.instrumentStorage.Remove(c, c.Param("symbol"))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	r.log.Info("Removed instrument", zap.String("symbol", c.Param("symbol")))
	r.stream.Publish("instrument.removed", []byte(c.Param("symbol")))
	c.JSON(200, gin.H{
		"message": "ok",
	})
}
