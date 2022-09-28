package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "YC Serverless"
	app.Usage = "Yandex Cloud (R) Serverless Containers (c) deployment plugin"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:      "key-file",
			Usage:     "file containing service account private key (json)",
			TakesFile: true,
		},
		cli.StringFlag{
			Name:   "key",
			EnvVar: "PLUGIN_KEY",
			Usage:  "service account private key contents (json)",
		},
		cli.StringFlag{
			Name:     "container-id",
			Usage:    "id of the container to deploy the new revision for",
			EnvVar:   "PLUGIN_CONTAINER_ID",
			Required: true,
		},
	}

	if err := app.Run(os.Args); err != nil {
		var l log.FieldLogger = log.StandardLogger()
		if e, ok := err.(HasFields); ok {
			l = log.WithFields(e.GetFields())
		}
		l.Fatal(err)
	}
}

func run(c *cli.Context) error {
	p := Plugin{
		Key:         []byte(c.String("key")),
		ContainerId: c.String("container-id"),
	}

	if keyFile := c.String("key-file"); keyFile != "" {
		if len(p.Key) != 0 {
			return fmt.Errorf("must not specify both --key and --key-file at the same time")
		}

		Info("reading service account key", log.Fields{"file": keyFile})

		f, err := os.Open(keyFile)
		if err != nil {
			return WithFields(fmt.Errorf("failed to open key file: %w", err), log.Fields{
				"file": keyFile,
			})
		}

		if err = func() error {
			defer f.Close()

			if p.Key, err = io.ReadAll(f); err != nil {
				return WithFields(fmt.Errorf("failed to read key file: %w", err), log.Fields{
					"file": keyFile,
				})
			}

			return nil
		}(); err != nil {
			return err
		}
	} else if len(p.Key) == 0 {
		return fmt.Errorf("must specify service account key either in --key or --key-file")
	} else {
		Info("got service account key from args")
	}

	return p.Exec()
}
