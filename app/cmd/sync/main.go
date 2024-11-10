package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"

	"github.com/sisisin/gh-project-sync/lib"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

var isLocal = os.Getenv("IS_LOCAL") == "true"
var token string = os.Getenv("GITHUB_TOKEN")
var org string = os.Getenv("GITHUB_ORG")
var ghProjectNumbers []int
var projectIDFromEnv string = os.Getenv("PROJECT_ID")
var cloudRunTaskIndex int

var dryRun *bool
var verbose *bool

func init() {
	if token == "" {
		panic("GITHUB_TOKEN is not set")
	}

	if org == "" {
		panic("GITHUB_ORG is not set")
	}

	{
		projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBERS")
		if projectNumberStr == "" {
			panic("GITHUB_PROJECT_NUMBER is not set")
		}
		for _, str := range strings.Split(projectNumberStr, ",") {
			num, err := strconv.Atoi(str)
			if err != nil {
				panic(err)
			}
			ghProjectNumbers = append(ghProjectNumbers, num)
		}
	}

	{
		iStr := os.Getenv("CLOUD_RUN_TASK_INDEX")
		if !isLocal && iStr == "" {
			panic("CLOUD_RUN_TASK_INDEX is not set")
		}
		var err error
		cloudRunTaskIndex, err = strconv.Atoi(iStr)
		if err != nil && !isLocal {
			panic(err)
		}

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

	graphqlClient := lib.NewGithubClient(token)

	var targetProjectNumber int
	if isLocal {
		logger.Infof(ctx, "local mode run. target project number is the first one: %d", ghProjectNumbers[0])
		targetProjectNumber = ghProjectNumbers[0]
	} else {
		logger.Infof(ctx, "target project number is %d", ghProjectNumbers[cloudRunTaskIndex])
		targetProjectNumber = ghProjectNumbers[cloudRunTaskIndex]
	}
	if err := lib.SyncProject(ctx, graphqlClient, org, targetProjectNumber); err != nil {
		logger.Error(ctx, "failed to run", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info(ctx, "end")
}
