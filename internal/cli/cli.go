package cli

import (
	"context"
	"os"
	"unicode"

	"github.com/gota33/initializr"
	"github.com/sirupsen/logrus"
	. "github.com/urfave/cli/v2"
	initmysql "server/internal/cli/config/mysql/v1"
	"server/internal/server"
)

const EnvPrefix = "APP_"

var (
	AppName = "app"
	Version = "dev"

	flagLevel     = flagName[string]("level")
	flagHttp      = flagName[string]("http")
	flagConfigUrl = flagName[string]("config-url")

	cli = &App{
		Name:    AppName,
		Version: Version,
		Flags: []Flag{
			&StringFlag{
				Name:    string(flagLevel),
				EnvVars: flagLevel.Envs(),
				Value:   "info",
			},
		},
		Before: func(c *Context) (err error) {
			var lvl logrus.Level
			if lvl, err = logrus.ParseLevel(flagLevel.Get(c)); err != nil {
				return
			}
			logrus.SetLevel(lvl)
			return
		},
		Commands: []*Command{
			{
				Name: "server",
				Flags: []Flag{
					&StringFlag{
						Name:    string(flagHttp),
						EnvVars: flagHttp.Envs(),
						Value:   ":8080",
					},
					&StringFlag{
						Name:    string(flagConfigUrl),
						EnvVars: flagConfigUrl.Envs(),
						Value:   "",
					},
				},
				Action: runServer,
			},
		},
	}
)

type flagName[T any] string

func (name flagName[T]) Get(c *Context) T {
	return c.Value(string(name)).(T)
}

func (name flagName[T]) Envs() []string {
	chars := []rune(EnvPrefix + name)
	for i, c := range chars {
		if c == '-' {
			chars[i] = '_'
		} else {
			chars[i] = unicode.ToUpper(c)
		}
	}
	return []string{string(chars)}
}

func runServer(c *Context) (err error) {
	var (
		config   server.Config
		res      initializr.Resource
		closeRDS func()
	)
	if configUrl := flagConfigUrl.Get(c); configUrl != "" {
		if res, err = initializr.FromJsonRemote(configUrl); err != nil {
			return
		}
		if config.RDS, closeRDS, err = initmysql.New(res, "rds"); err != nil {
			return
		}
		defer closeRDS()
	}
	config.Addr = flagHttp.Get(c)
	return server.Run(c.Context, config)
}

func Run(ctx context.Context) (err error) {
	return cli.RunContext(ctx, os.Args)
}
