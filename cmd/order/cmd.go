package main

import (
	"context"
	"errors"
	"github.com/klwxsrx/arch-course-project/data/mysql/order"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	loggerImpl "github.com/klwxsrx/arch-course-project/pkg/common/infra/logger"
	commonMysql "github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/deliveryapi"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/paymentapi"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/warehouseapi"
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

	db, client, err := getDatabaseClient(config, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup db connection")
	}
	defer db.Close()

	migration, err := commonMysql.NewMigration(client, logger, order.MysqlMigrations)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup db migration")
	}
	err = migration.Migrate()
	if err != nil {
		logger.WithError(err).Fatal("failed to execute db migration")
	}

	paymentAPI := paymentapi.New(config.PaymentServiceURL)
	warehouseAPI := warehouseapi.New(config.WarehouseServiceURL)
	deliveryAPI := deliveryapi.New(config.DeliveryServiceURL)

	unitOfWork := mysql.NewUnitOfWork(client)
	orderService := service.NewOrderService(
		paymentAPI,
		warehouseAPI,
		deliveryAPI,
		unitOfWork,
		logger,
	)

	server, err := startServer(orderService, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to start server")
	}
	logger.Info("app is ready")

	listenOSKillSignals()
	_ = server.Shutdown(context.Background())
}

func getDatabaseClient(config *config, logger log.Logger) (commonMysql.Connection, commonMysql.TransactionalClient, error) {
	db, err := commonMysql.NewConnection(commonMysql.Config{DSN: commonMysql.Dsn{
		User:     config.DBUser,
		Password: config.DBPassword,
		Host:     config.DBHost,
		Port:     config.DBPort,
		Database: config.DBName,
	}}, logger)
	if err != nil {
		return nil, nil, err
	}
	client, err := db.Client()
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	return db, client, nil
}

func startServer(orderService *service.OrderService, logger log.Logger) (*http.Server, error) {
	handler, err := transport.NewHTTPHandler(orderService, logger)
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