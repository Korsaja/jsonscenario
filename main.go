package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/Korsaja/jsonscenario/internal/pipeline"

	"github.com/urfave/cli/v2"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout *os.File, stderr *os.File) int {
	logger := log.New(stderr, "[LOG] ", log.Ldate|log.Ltime)
	app := &cli.App{
		Name:      "app",
		Usage:     "run operation from config.json scenarios",
		Suggest:   true,
		Writer:    stdout,
		ErrWriter: stderr,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "scenario",
				Aliases:   []string{"s"},
				Usage:     "path to scenario",
				Required:  true,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "path to save result",
				Value:   "result.json",
			},
		},
		Action: func(c *cli.Context) error {
			scenario := c.String("scenario")
			output := c.String("out")

			if ext := filepath.Ext(scenario); ext != ".json" {
				return errors.New("scenario must have json ext")
			}

			pipe, err := pipeline.NewPipeline(scenario, output, logger)
			if err != nil {
				return err
			}
			return pipe.Do()
		},
	}

	if err := app.Run(args); err != nil {
		logger.Printf("app failed: %s", err.Error())
		return 1
	}

	return 0
}
