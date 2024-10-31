package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "redirector",
		Usage:                "Interact with URL Redirect API",
		Version:              "v1.3",
		Compiled:             time.Now(),
		EnableBashCompletion: true,
		Args:                 true,
		Authors: []*cli.Author{
			{
				Name:  "Kushwanth Reddy",
				Email: "human@gkr.dev",
			},
		},
		Action: getStatus,
		Commands: []*cli.Command{
			{
				Name:               "status",
				Usage:              "check api host status",
				Args:               false,
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             getStatus,
			},
			{
				Name:               "get",
				Usage:              "get an existing redirect",
				Args:               true,
				ArgsUsage:          "id",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             getUrlRedirect,
			},
			{
				Name:               "create",
				Usage:              "create an new redirect",
				Args:               true,
				ArgsUsage:          "path url",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             addUrlRedirect,
			},
			{
				Name:               "update",
				Usage:              "update an existing redirect",
				Args:               true,
				ArgsUsage:          "id path url",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             updateUrlRedirect,
			},
			{
				Name:               "delete",
				Usage:              "disable an existing redirect",
				Args:               true,
				ArgsUsage:          "id",
				CustomHelpTemplate: commandHelpText,
				Action:             disableUrlRedirect,
			},
			{
				Name:               "fix",
				Usage:              "fix an existing redirect",
				Args:               true,
				ArgsUsage:          "path url",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             fixUrlRedirect,
			},
			{
				Name:               "list",
				Usage:              "list all redirects",
				Args:               true,
				ArgsUsage:          "page",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             listUrlRedirects,
			},
			{
				Name:               "search",
				Usage:              "search for a redirect",
				Args:               true,
				ArgsUsage:          "search_text page",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             searchUrlRedirect,
			},
			{
				Name:               "check",
				Usage:              "check if a redirect exists",
				Args:               true,
				ArgsUsage:          "url",
				HideHelpCommand:    true,
				CustomHelpTemplate: commandHelpText,
				Action:             urlRedirectExists,
			},
		},
		CustomAppHelpTemplate: appHelpText,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		defer os.Exit(1)
	}
}
