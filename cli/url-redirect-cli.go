package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if !isAPIUp() {
		fmt.Print("API Server is down")
		defer os.Exit(1)
	}
	action := os.Args[1]
	switch {
	case action == "info":
		infoFlags := flag.NewFlagSet("info", flag.ExitOnError)
		id := infoFlags.Int("id", 0, "Id of Redirect")
		infoFlags.Parse(os.Args[2:])
		getRedirect(*id)
	case action == "create":
		createFlags := flag.NewFlagSet("create", flag.ExitOnError)
		path := createFlags.String("path", "", "Custom Path")
		uri := os.Args[4]
		createFlags.Parse(os.Args[2:])
		createRedirect(uri, *path)
	case action == "update":
		updateFlags := flag.NewFlagSet("update", flag.ExitOnError)
		id := updateFlags.Int("id", 0, "Id of Redirect")
		path := updateFlags.String("path", "", "Custom Path")
		uri := os.Args[6]
		updateFlags.Parse(os.Args[2:])
		updateRedirect(uri, *id, *path)
	case action == "fix":
		fixFlags := flag.NewFlagSet("fix", flag.ExitOnError)
		path := fixFlags.String("path", "", "Custom Path")
		fixFlags.Parse(os.Args[2:])
		uri := os.Args[4]
		fixRedirect(uri, *path)
	case action == "delete":
		deleteFlags := flag.NewFlagSet("delete", flag.ExitOnError)
		id := deleteFlags.Int("id", 0, "Id of Redirect")
		deleteFlags.Parse(os.Args[2:])
		deleteRedirect(*id)
	case action == "status":
		if isAPIUp() {
			fmt.Println("UP")
		} else {
			fmt.Println("DOWN")
		}
	default:
		fmt.Print("CLI to interact with URL Redirect API\n\n")
	}
}
