package marketdata

import (
	"github.com/alpacahq/alpaca-trade-api-go/v2/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v2/marketdata"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(routeGroup *gin.RouterGroup, log *zap.Logger, md marketdata.Client, trade alpaca.Client, ohlcStorage OHLCStorage) {
	r := &Routes{log: log, marketdata: md, trade: trade, ohlcStorage: ohlcStorage}
	routeGroup.GET("/:symbol", r.GetMarketData)
}

type Routes struct {
	log         *zap.Logger
	marketdata  marketdata.Client
	trade       alpaca.Client
	ohlcStorage OHLCStorage
}

// GetMarketData godoc
//
//	@Summary	Get OHLC data for an Instrument
//	@Tags		marketdata
//	@Accept		json
//	@Produce	json
//	@Param		symbol	path		string	true	"Instrument Symbol"
//	@Success	200		{object}	OHLC
//	@Router		/marketdata/{symbol} [get]
func (r *Routes) GetMarketData(c *gin.Context) {
	ohlc, err := r.ohlcStorage.Get(c, c.Param("symbol"))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, ohlc)
}
