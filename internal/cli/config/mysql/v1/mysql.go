package v1

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gota33/initializr"
	"github.com/sirupsen/logrus"
)

type MySQLOptions struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Database string   `json:"database"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	MaxOpen  int      `json:"maxOpen"`
	MaxIdle  int      `json:"maxIdle"`
	Params   []string `json:"params"`
}

func New(res initializr.Resource, key string) (db *sql.DB, close func(), err error) {
	var opts MySQLOptions
	if err = res.Scan(key, &opts); err != nil {
		return
	}

	params := make(map[string]string, len(opts.Params))
	for _, param := range opts.Params {
		kv := strings.SplitN(param, "=", 2)
		params[kv[0]] = kv[1]
	}

	c := mysql.NewConfig()
	c.User = opts.Username
	c.Passwd = opts.Password
	c.Net = "tcp"
	c.Addr = opts.Host + ":" + opts.Port
	c.DBName = opts.Database
	c.Params = params

	if db, err = sql.Open("mysql", c.FormatDSN()); err != nil {
		return
	}

	if opts.MaxOpen > 0 {
		db.SetMaxOpenConns(opts.MaxOpen)
	}
	if opts.MaxIdle > 0 {
		db.SetMaxIdleConns(opts.MaxIdle)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return
	}

	close = func() {
		if closeErr := db.Close(); closeErr != nil {
			logrus.WithError(err).Warnf("Close MySQL error")
		}
	}
	return
}
