package main

import (
	"context"
	"net/http"

	"github.com/alpacahq/alpaca-trade-api-go/v2/alpaca"
	alpacamd "github.com/alpacahq/alpaca-trade-api-go/v2/marketdata"
	"github.com/gin-gonic/gin"
	_ "github.com/lhjnilsson/go-fx-gin-nats-pgx/docs"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/instrument"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/internal/config"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/internal/database/postgres"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/internal/logging"
	"github.com/lhjnilsson/go-fx-gin-nats-pgx/marketdata"
	"github.com/nats-io/nats.go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// @title			Go FX Gin NATS PGX
// @version		1.0
// @host		localhost:8080
func NewHTTPServer(lc fx.Lifecycle, log *zap.Logger, config *config.ApplicationConfig) *gin.Engine {
	gin := gin.New()
	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: gin,
	}
	log.Info("HTTP server configured", zap.String("address", config.ServerAddress))
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting HTTP server")
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Fatal("HTTP server failed", zap.Error(err))
				}
			}()
			log.Info("HTTP server started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
	return gin
}

func NewMarketdataClient(config *config.ApplicationConfig) (alpacamd.Client, error) {
	md := alpacamd.NewClient(alpacamd.ClientOpts{ApiKey: config.AlpacaAPIKey, ApiSecret: config.AlpacaAPISecret})
	_, err := md.GetLatestBar("AAPL")
	return md, err
}

func NewAlpacaClient(config *config.ApplicationConfig) (alpaca.Client, error) {
	cl := alpaca.NewClient(alpaca.ClientOpts{ApiKey: config.AlpacaAPIKey, ApiSecret: config.AlpacaAPISecret, BaseURL: config.AlpacaAPIURL})
	_, err := cl.GetAccount()
	return cl, err
}

func NewStreamingClient(config *config.ApplicationConfig) (nats.JetStream, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return js, nil
}

func RegisterRoutes(gin *gin.Engine, log *zap.Logger, md alpacamd.Client, trade alpaca.Client, instruments instrument.InstrumentStorage, ohlc marketdata.OHLCStorage, stream nats.JetStream) error {
	group := gin.Group("/instruments")
	instrument.RegisterRoutes(group, log, md, trade, instruments, stream)
	group = gin.Group("/marketdata")
	marketdata.RegisterRoutes(group, log, md, trade, ohlc)
	gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Info("Routes registered")
	return nil
}

func main() {
	fx.New(
		fx.Provide(
			context.Background,
			config.New,
			logging.New,
			postgres.New,
			NewStreamingClient,
			instrument.NewPostgresStorage,
			instrument.NewInstrumentStream,
			marketdata.NewPostgresStorage,
			NewAlpacaClient,
			NewMarketdataClient,
			NewHTTPServer,
		),
		fx.Invoke(
			RegisterRoutes,
			marketdata.NewMarketDataStreamer,
		),
	).Run()
}
