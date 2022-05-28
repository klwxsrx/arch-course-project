package main

import (
	"context"
	"errors"
	migrations "github.com/klwxsrx/arch-course-project/data/mysql/auth"
	"github.com/klwxsrx/arch-course-project/pkg/auth/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/auth"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/encoding"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/redis"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	loggerImpl "github.com/klwxsrx/arch-course-project/pkg/common/infra/logger"
	commonMysql "github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
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

	db, client, err := getDatabaseClient(config, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup db connection")
	}
	defer db.Close()

	migration, err := commonMysql.NewMigration(client, logger, migrations.MysqlMigrations)
	if err != nil {
		logger.WithError(err).Fatal("failed to setup db migration")
	}
	err = migration.Migrate()
	if err != nil {
		logger.WithError(err).Fatal("failed to execute db migration")
	}

	redisCli, err := commonRedis.NewClient(&commonRedis.Config{
		Address:  config.RedisAddress,
		Password: config.RedisPassword,
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to setup redis connection")
	}
	defer redisCli.Close()

	userRepo := mysql.NewUserRepository(client)
	passwordEncoder := encoding.NewPasswordEncoder()
	userService := service.NewUserService(userRepo, passwordEncoder)
	sessionService := auth.NewSessionService(redis.NewSessionStorage(redisCli), userRepo, passwordEncoder)

	server, err := startServer(userService, sessionService, logger)
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

func startServer(userService *service.UserService, sessionService *auth.SessionService, logger log.Logger) (*http.Server, error) {
	handler, err := transport.NewHTTPHandler(userService, sessionService, logger)
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
