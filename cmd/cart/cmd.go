package main

import (
	"context"
	"errors"
	"github.com/klwxsrx/arch-course-project/pkg/cart/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/cart/infra/catalogapi"
	"github.com/klwxsrx/arch-course-project/pkg/cart/infra/orderapi"
	"github.com/klwxsrx/arch-course-project/pkg/cart/infra/redis"
	"github.com/klwxsrx/arch-course-project/pkg/cart/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	loggerImpl "github.com/klwxsrx/arch-course-project/pkg/common/infra/logger"
	commonRedis "github.com/klwxsrx/arch-course-project/pkg/common/infra/redis"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := loggerImpl.New()

	config, err := parseConfig()
	if err != nil {
		logger.WithError(err).Fatal("failed to parse config")
	}

	redisCli, err := commonRedis.NewClient(&commonRedis.Config{
		Address:  config.RedisAddress,
		Password: config.RedisPassword,
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to setup redis connection")
	}
	defer redisCli.Close()

	cartService := service.NewCartService(
		catalogapi.New(config.CatalogServiceURL),
		orderapi.New(config.OrderServiceURL),
		redis.NewCartStorage(redisCli),
		logger,
	)

	server, err := startServer(cartService, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
	logger.Info("app is ready")

	listenOSKillSignals()
	_ = server.Shutdown(context.Background())
}

func startServer(service *service.CartService, logger log.Logger) (*http.Server, error) {
	handler, err := transport.NewHTTPHandler(service, logger)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("unable to start the server")
		}
	}()
	return srv, nil
}

func listenOSKillSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}
