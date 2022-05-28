package mysql

import (
	"fmt"
	"github.com/cenkalti/backoff"
	_ "github.com/go-sql-driver/mysql" // driver impl
	"github.com/jmoiron/sqlx"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"time"
)

const (
	maxOpenConnections = 10
	connTimeout        = 30 * time.Second
)

type Config struct {
	DSN Dsn
}

type Dsn struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func (d Dsn) String() string {
	return fmt.Sprintf("%s:%s@(%s:%s)/%s", d.User, d.Password, d.Host, d.Port, d.Database)
}

type Connection interface {
	Client() (TransactionalClient, error)
	Close()
}

type connection struct {
	config Config
	db     *sqlx.DB
	logger log.Logger
}

func (c *connection) Client() (TransactionalClient, error) {
	return &client{c.db}, nil
}

func (c *connection) Close() {
	err := c.db.Close()
	if err != nil {
		c.logger.WithError(err).Error("failed to close mongo db connection")
	}
}

func (c *connection) openConnection() error {
	var err error
	c.db, err = sqlx.Open("mysql", c.config.DSN.String()+"?parseTime=true")
	if err != nil {
		return err
	}

	c.db.SetMaxOpenConns(maxOpenConnections)

	err = backoff.Retry(func() error {
		return c.db.Ping()
	}, newOpenConnectionBackoff(connTimeout))
	if err != nil {
		_ = c.db.Close()
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}
	return nil
}

func newOpenConnectionBackoff(connTimeout time.Duration) *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = connTimeout
	return b
}

func NewConnection(config Config, logger log.Logger) (Connection, error) {
	db := &connection{config: config, logger: logger}
	err := db.openConnection()
	return db, err
}
