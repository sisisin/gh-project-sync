package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/compute/metadata"

	"github.com/sisisin/gh-project-sync/lib"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

var isLocal = os.Getenv("IS_LOCAL") == "true"
var token string = os.Getenv("GITHUB_TOKEN")
var org string = os.Getenv("GITHUB_ORG")
var ghProjectNumber int
var projectIDFromEnv string = os.Getenv("PROJECT_ID")

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
	ghProjectNumber, err = strconv.Atoi(projectNumberStr)
	if err != nil {
		panic(err)
	}

	if !isLocal && projectIDFromEnv != "" {
		panic("PROJECT_ID must only set in local")
	}

	dryRun = flag.Bool("dry-run", false, "dry run")
	verbose = flag.Bool("verbose", false, "verbose")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	logger.SetDefaultLogger()
	var projectID string
	if isLocal {
		projectID = projectIDFromEnv
	} else {
		var err error
		projectID, err = metadata.ProjectIDWithContext(ctx)
		if err != nil {
			panic(err)
		}
	}

	ctx = appcontext.WithDryRun(ctx, *dryRun)
	ctx = appcontext.WithVerbose(ctx, *verbose)
	ctx = appcontext.WithProjectID(ctx, projectID)
	ctx = appcontext.WithTraceID(ctx, time.Now().Format("20060102150405"))

	logger.Info(ctx, "start")

	if err := run(ctx); err != nil {
		logger.Error(ctx, "failed to run", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info(ctx, "end")
}

func run(ctx context.Context) error {
	graphqlClient := lib.NewGithubClient(token)

	err := lib.GetProjectDetailAll(ctx, graphqlClient, org, ghProjectNumber)
	if err != nil {
		return err
	}
	return nil
}
