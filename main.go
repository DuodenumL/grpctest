package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/projecteru2/grpctest/pbreflect"
	"github.com/projecteru2/grpctest/testsuite"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Usage = "test a grpc service"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "proto",
			Aliases: []string{"p"},
			Usage:   "proto file pathname",
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "grpc service address, e.g. localhost:5001",
		},
		&cli.StringFlag{
			Name:    "testsuite",
			Aliases: []string{"t"},
			Usage:   "testsuite yaml pathname",
		},
	}
	app.Action = action
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}

func action(c *cli.Context) (err error) {
	service, err := pbreflect.Parse(c.String("proto"))
	if err != nil {
		return
	}
	if err = service.SetAddress(c.String("address")); err != nil {
		return
	}

	stdout, stderr, err := testsuite.Preprocess(c.String("testsuite"))
	if err != nil {
		var errMsg []byte
		if stderr != nil {
			errMsg, _ = ioutil.ReadAll(stderr)
		}
		return errors.WithMessage(err, string(errMsg))
	}
	suites := testsuite.MustParse(stdout)

	t := &T{}
	for suite := range suites {
		if !suite.Run(t, service) {
			return errors.New("test failed")
		}
	}

	return
}
