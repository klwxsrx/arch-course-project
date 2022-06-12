package main

import (
	"context"
	"errors"
	"github.com/klwxsrx/arch-course-project/data/mysql/payment"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	commonMessage "github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	loggerImpl "github.com/klwxsrx/arch-course-project/pkg/common/infra/logger"
	commonMysql "github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/pulsar"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/payment/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/payment/infra/transport"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "payment"

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

	migration, err := commonMysql.NewMigration(client, logger, payment.MysqlMigrations)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup db migration")
	}
	err = migration.Migrate()
	if err != nil {
		logger.WithError(err).Fatal("failed to execute db migration")
	}

	pulsarConn, err := pulsar.NewConnection(config.MessageBrokerAddress, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup message broker connection")
	}

	messageSender := pulsar.NewMessageSender(pulsarConn)
	defer messageSender.Close()

	messageDispatcher := commonMessage.NewDispatcher(
		commonMysql.NewMessageStore(client),
		messageSender,
		commonMysql.NewSynchronization(client),
		logger,
	)
	defer messageDispatcher.Close()
	messageDispatcher.Dispatch()

	unitOfWork := mysql.NewUnitOfWork(client)
	unitOfWork = persistence.NewUnitOfWorkCompleteNotifier(unitOfWork, messageDispatcher.Dispatch)
	paymentService := service.NewPaymentService(
		unitOfWork,
		logger,
	)
	paymentQueryService := mysql.NewPaymentQueryService(client)

	subscriberCloser, err := pulsar.NewMessageSubscriber(
		serviceName,
		[]commonMessage.Handler{
			message.NewAuthorizePaymentHandler(paymentService),
		},
		pulsarConn,
		logger,
	)
	if err != nil {
		logger.WithError(err).Fatal("failed to run message subscriber")
	}
	defer subscriberCloser()

	server, err := startServer(paymentService, paymentQueryService, logger)
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

func startServer(paymentService *service.PaymentService, paymentQueryService query.PaymentQueryService, logger log.Logger) (*http.Server, error) {
	handler, err := transport.NewHTTPHandler(paymentService, paymentQueryService, logger)
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
