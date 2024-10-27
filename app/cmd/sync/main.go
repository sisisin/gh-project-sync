package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/sisisin/gh-project-sync/lib"
)

var token string = os.Getenv("GITHUB_TOKEN")
var org string = os.Getenv("GITHUB_ORG")
var projectNumber int
var dryRun *bool
var verbose *bool

func init() {
	if token == "" {
		panic("GITHUB_TOKEN is not set")
	}

	if org == "" {
		panic("GITHUB_ORG is not set")
	}

	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		panic("GITHUB_PROJECT_NUMBER is not set")
	}
	var err error
	projectNumber, err = strconv.Atoi(projectNumberStr)
	if err != nil {
		panic(err)
	}

	dryRun = flag.Bool("dry-run", false, "dry run")
	verbose = flag.Bool("verbose", false, "verbose")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	ctx = lib.WithDryRun(ctx, *dryRun)
	ctx = lib.WithVerbose(ctx, *verbose)

	if err := run(ctx); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	graphqlClient := lib.NewGithubClient(token)

	err := lib.GetProjectDetailAll(ctx, graphqlClient, org, projectNumber)
	if err != nil {
		return err
	}
	return nil
}
