package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "redirector",
		Usage:                "Interact with URL Redirect API",
		Version:              "v1.2",
		Compiled:             time.Now(),
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Kushwanth Reddy",
				Email: "human@gkr.dev",
			},
		},
		Action: getStatus,
		Commands: []*cli.Command{
			{
				Name:   "status",
				Usage:  "check api host status",
				Action: getStatus,
			},
			{
				Name:      "get",
				Usage:     "get an existing redirect",
				ArgsUsage: "id",
				Action:    getUrlRedirect,
			},
			{
				Name:      "create",
				Usage:     "create an new redirect",
				ArgsUsage: "path url",
				Action:    addUrlRedirect,
			},
			{
				Name:      "update",
				Usage:     "update an existing redirect",
				ArgsUsage: "id path url",
				Action:    updateUrlRedirect,
			},
			{
				Name:      "delete",
				Usage:     "disable an existing redirect",
				ArgsUsage: "id",
				Action:    disableUrlRedirect,
			},
			{
				Name:      "fix",
				Usage:     "fix an existing redirect",
				ArgsUsage: "path url",
				Action:    fixUrlRedirect,
			},
			{
				Name:      "list",
				Usage:     "list all redirects",
				ArgsUsage: "page",
				Action:    listUrlRedirects,
			},
			{
				Name:      "search",
				Usage:     "search for a redirect",
				ArgsUsage: "search",
				Action:    searchUrlRedirect,
			},
			{
				Name:      "check",
				Usage:     "check if a redirect exists",
				ArgsUsage: "check",
				Action:    urlRedirectExists,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
