package v1

import (
	"context"
	"database/sql"
	"time"

	"github.com/gota33/initializr"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Options struct {
	DSN string `json:"dsn"`
}

func New(res initializr.Resource, key string) (db *sql.DB, close func(), err error) {
	var opts Options
	if err = res.Scan(key, &opts); err != nil {
		return
	}
	if db, err = sql.Open("sqlite3", opts.DSN); err != nil {
		return
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return
	}

	close = func() {
		if closeErr := db.Close(); closeErr != nil {
			logrus.WithError(err).Warnf("Close SQLite error")
		}
	}
	return
}
