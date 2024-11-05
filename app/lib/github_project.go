package lib

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/gcp/bigquery"
	"github.com/sisisin/gh-project-sync/lib/gcp/storage"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

func SyncProject(
	ctx context.Context,
	graphqlClient *GitHubClient,
	org string,
	projectNumber int,
) error {
	if err := WriteToGcs(ctx); err != nil {
		return errors.Wrap(err, "failed to write to GCS")
	}
	if err := LoadToBigQuery(ctx); err != nil {
		return errors.Wrap(err, "failed to load to BigQuery")
	}

	return nil
}

func LoadToBigQuery(ctx context.Context) error {
	logger.Info(ctx, "start loading to BigQuery")
	// TODO: fix hardcoded datasetId and tableId
	table, err := bigquery.New(ctx, "github_project_sync", "project_items")
	if err != nil {
		return errors.Wrap(err, "failed to create table")
	}

	if err := table.DeleteByProjectNumber(ctx, "2024-10-31T16:00:00", 9); err != nil {
		return errors.Wrap(err, "failed to delete")
	}
	if err := table.Load(ctx, "gs://github-project-sync/test1.ndjson", "2024103116"); err != nil {
		return errors.Wrap(err, "failed to load")
	}

	logger.Info(ctx, "end loading to BigQuery")
	return nil
}

func WriteToGcs(ctx context.Context) error {
	logger.Info(ctx, "start writing to GCS")
	appStorage, err := storage.New(ctx, "github-project-sync")
	if err != nil {
		return errors.Wrap(err, "failed to create storage")
	}

	values := []string{
		`{"organization_id":"knowledge-work", "project_number":10,creator:"sisisin"}`,
		`{"organization_id":"knowledge-work", "project_number":10,creator:"foobar"}`,
	}
	objectWriter := appStorage.GetObjectWriter(ctx, "test3.ndjson")
	objectWriter.ContentType = "application/x-ndjson"

	for i, v := range values {
		line := v
		if i != len(values)-1 {
			line += "\n"
		}
		if _, err := objectWriter.Write([]byte(line)); err != nil {
			return errors.Wrap(err, "failed to write object")
		}
	}
	if err := objectWriter.Close(); err != nil {
		return errors.Wrap(err, "failed to close object")
	}

	logger.Info(ctx, "end writing to GCS")
	return nil
}

func GetProjectDetailAll(
	ctx context.Context,
	graphqlClient *GitHubClient,
	org string,
	projectNumber int,
) error {

	projectId, err := graphqlClient.GetProjectID(ctx, org, projectNumber)
	if err != nil {
		return errors.Wrap(err, "failed to query GetProjectID")
	}

	projectDetail, err := graphqlClient.GetProjectDetail(ctx, projectId)
	if err != nil {
		return errors.Wrap(err, "failed to query GetProjectDetail")
	}

	if appcontext.GetVerbose(ctx) {
		logger.Info(ctx, "rateLimit", slog.Any("body", projectDetail["rateLimit"]))
	}

	j, err := json.Marshal(projectDetail)
	if err != nil {
		return errors.Wrap(err, "failed to marshal projectDetail")
	}

	logger.Info(ctx, "response", slog.String("body", string(j)))

	return nil
}
