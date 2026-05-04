package main

import (
	"code/crawler"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	cli "github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Usage: "analyze a website structure",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "depth",
				Value: 10,
				Usage: "crawl depth",
			},
			&cli.UintFlag{
				Name:  "retries",
				Value: 1,
				Usage: "number of retries for failed requests",
			},
			&cli.DurationFlag{
				Name:  "delay",
				Value: 0,
				Usage: "delay between requests (example: 200ms, 1s)",
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Value: 15 * time.Second,
				Usage: "per-request timeout",
			},
			&cli.UintFlag{
				Name:  "rps",
				Value: 0,
				Usage: "limit requests per second (overrides delay)",
			},
			&cli.StringFlag{
				Name:  "user-agent",
				Usage: "custom user agent",
			},
			&cli.UintFlag{
				Name:  "workers",
				Value: 4,
				Usage: "number of concurrent workers",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "url",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Shows a list of commands or help for one command",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return cli.ShowAppHelp(cmd)
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			url := cmd.StringArg("url")

			if url == "" {
				return cli.ShowAppHelp(cmd)
			}

			var opts crawler.Options

			opts.URL = url
			opts.Depth = cmd.Uint("depth")
			opts.Retries = cmd.Uint("retries")
			opts.Delay = cmd.Duration("delay")
			opts.Timeout = cmd.Duration("timeout")
			opts.UserAgent = cmd.String("user-agent")
			opts.Concurrency = cmd.Uint("workers")
			opts.IndentJSON = 2
			opts.HTTPClient = &http.Client{}

			result, err := crawler.Analize(ctx, opts)
			if err != nil {
				return err
			}

			fmt.Println(string(result))

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
