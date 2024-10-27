package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/urfave/cli/v2"
)

func getStatus(*cli.Context) error {
	if isAPIUp() {
		fmt.Println("API is running")
	} else {
		fmt.Println("API is Down")
	}
	return nil
}

func getUrlRedirect(cCtx *cli.Context) error {
	idStr := cCtx.Args().Get(0)
	id, idErr := strconv.Atoi(idStr)
	if idErr != nil {
		respondAndExit("Args Error", idErr)
	}
	getRedirect(id)
	return nil
}

func addUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	createRedirect(uri, path)
	return nil
}

func updateUrlRedirect(cCtx *cli.Context) error {
	idStr, path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1), cCtx.Args().Get(2)
	id, idErr := strconv.Atoi(idStr)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if idErr != nil || pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", idErr, pathErr, uriErr)
	}
	updateRedirect(id, path, uri)
	return nil
}

func fixUrlRedirect(cCtx *cli.Context) error {
	path, uri := cCtx.Args().Get(0), cCtx.Args().Get(1)
	_, pathErr := url.Parse(path)
	_, uriErr := url.Parse(uri)
	if pathErr != nil || uriErr != nil {
		respondAndExit("Args Error", pathErr, uriErr)
	}
	fixRedirect(uri, path)
	return nil
}

func disableUrlRedirect(cCtx *cli.Context) error {
	idStr := cCtx.Args().Get(0)
	id, idErr := strconv.Atoi(idStr)
	if idErr != nil {
		respondAndExit("Args Error", idErr)
	}
	deleteRedirect(id)
	return nil
}

func listUrlRedirects(cCtx *cli.Context) error {
	pageStr := cCtx.Args().Get(0)
	page, pageErr := strconv.Atoi(pageStr)
	if pageErr != nil {
		page = 0
	} else {
		page = page * 10
	}
	listRedirect(page)
	return nil
}

func searchUrlRedirect(cCtx *cli.Context) error {
	path, pageStr := cCtx.Args().Get(0), cCtx.Args().Get(1)
	page, pageErr := strconv.Atoi(pageStr)
	_, pathErr := url.Parse(path)
	if pathErr != nil {
		respondAndExit("Args Error", pathErr)
	}
	if pageErr != nil {
		page = 0
	} else {
		page = page * 10
	}
	searchRedirect(path, page)
	return nil
}

func urlRedirectExists(cCtx *cli.Context) error {
	path := cCtx.Args().Get(0)
	_, pathErr := url.Parse(path)
	if pathErr != nil {
		respondAndExit("Args Error", pathErr)
	}
	redirectExists(path)
	return nil
}
