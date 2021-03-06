package main

import (
	"github.com/gota33/initializr"
	"github.com/sirupsen/logrus"
	"server/internal/cli"
)

func main() {
	ctx, cancel := initializr.GracefulContext()
	defer cancel()

	if err := cli.Run(ctx); err != nil {
		logrus.WithError(err).Fatal("Exit with error")
	}
}
