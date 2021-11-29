package postgres

import (
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // DB driver
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	maxConnectionRetry     = 60
	connectionRetryDelayMs = 1000
)

type Connector interface {
	Open(dsn DSN) error
	WaitUntilReady() error
	Ready() bool
	Client() *sqlx.DB
	Close() error
}

func NewConnector() Connector {
	return &connector{}
}

type connector struct {
	db    *sqlx.DB
	ready bool
}

func (c *connector) Open(dsn DSN) error {
	var err error
	c.db, err = sqlx.Open("pgx", dsn.String())
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	return nil
}

func (c *connector) WaitUntilReady() error {
	for i := 0; i < maxConnectionRetry; i++ {
		err := c.db.Ping()
		if err == nil {
			c.ready = true
			return nil
		}
		time.Sleep(time.Duration(connectionRetryDelayMs) * time.Millisecond)
	}
	return errors.New("failed to connect to database")
}

func (c *connector) Ready() bool {
	return c.ready
}

func (c *connector) Client() *sqlx.DB {
	if !c.ready {
		panic("db client not ready, but requested")
	}
	return c.db
}

func (c *connector) Close() error {
	err := c.db.Close()
	return errors.Wrap(err, "failed to disconnect")
}
