package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/pkg/errors"
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
	now := time.Now().UTC()
	if err := GetAndWriteToGcs(ctx, graphqlClient, now, org, projectNumber); err != nil {
		return errors.Wrap(err, "failed to write to GCS")
	}
	if err := LoadToBigQuery(ctx, now, projectNumber); err != nil {
		return errors.Wrap(err, "failed to load to BigQuery")
	}

	return nil
}

func LoadToBigQuery(ctx context.Context, now time.Time, projectNumber int) error {
	logger.Info(ctx, "start loading to BigQuery")
	// TODO: fix hardcoded datasetId and tableId
	table, err := bigquery.New(ctx, "github_project_sync", "project_items")
	if err != nil {
		return errors.Wrap(err, "failed to create table")
	}

	if err := table.DeleteByProjectNumber(ctx, now, projectNumber); err != nil {
		return errors.Wrap(err, "failed to delete")
	}
	// todo: fix hardcoded bucket name
	loadTarget := "gs://github-project-sync-knowledgework-simenyan-sandbox/" + getOutFilePathFromNow(now, projectNumber)
	if err := table.Load(ctx, loadTarget, now); err != nil {
		return errors.Wrap(err, "failed to load")
	}

	logger.Info(ctx, "end loading to BigQuery")
	return nil
}

func getOutFilePathFromNow(now time.Time, projectNumber int) string {
	return now.Format("2006-01-02/1504") + fmt.Sprintf("-project_%d-out.ndjson", projectNumber)
}

func GetAndWriteToGcs(ctx context.Context,
	graphqlClient *GitHubClient,
	now time.Time,
	org string,
	projectNumber int,
) error {
	logger.Info(ctx, "start get project info and writing to GCS")

	appStorage, err := storage.New(ctx, "github-project-sync-knowledgework-simenyan-sandbox")
	if err != nil {
		return errors.Wrap(err, "failed to create storage")
	}
	objectWriter := appStorage.GetObjectWriter(ctx, getOutFilePathFromNow(now, projectNumber))
	objectWriter.ContentType = "application/x-ndjson"

	projectSummary, err := graphqlClient.GetProjectSummary(ctx, org, projectNumber)
	if err != nil {
		return errors.Wrap(err, "failed to GetProjectSummary")
	}

	var (
		rateLimit  map[string]any
		totalCount int
		cursor     string
		gotItems   int
	)
	for {
		logger.Infof(ctx, "processing by project items %d - %d", gotItems, gotItems+100)
		items, err := graphqlClient.GetProjectItems(ctx, projectSummary.Organization.ProjectV2.Id, cursor)
		if err != nil {
			return errors.Wrap(err, "failed to get project items")
		}
		gotItems += len(items.Node.Items.Nodes)

		jsonLines, err := toJsonLines(projectSummary, items, items.Node.Items.PageInfo.HasNextPage)
		if err != nil {
			return errors.Wrap(err, "failed to toJsonLines")
		}
		if _, err := objectWriter.Write(jsonLines); err != nil {
			return errors.Wrap(err, "failed to write object")
		}

		if items.Node.Items.PageInfo.HasNextPage {
			cursor = items.Node.Items.PageInfo.EndCursor
		} else {
			rateLimit = items.RateLimit
			totalCount = items.Node.Items.TotalCount
			break
		}
	}

	// TODO: log rateLimit's cost for all requests
	logger.Info(ctx, "end get project info and writing to GCS",
		slog.Any("rateLimit", rateLimit),
		slog.Any("gotItems", gotItems),
		slog.Any("totalCount", totalCount),
	)

	if err := objectWriter.Close(); err != nil {
		return errors.Wrap(err, "failed to close object")
	}

	logger.Info(ctx, "end writing to GCS")
	return nil
}

type RFC3339Time struct {
	time.Time
}

func (b RFC3339Time) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", b.Format(time.RFC3339))
	return []byte(formatted), nil
}

type OutJson struct {
	OrganizationLogin string         `json:"organization_login"`
	ProjectNumber     int            `json:"project_number"`
	ProjectTitle      string         `json:"project_title"`
	Item              map[string]any `json:"item"`
}

func toJsonLines(summary *GitHubProjectSummary, items *GitHubProjectItems, hasNextPage bool) ([]byte, error) {
	var lines []byte
	for i, v := range items.Node.Items.Nodes {
		value := OutJson{
			OrganizationLogin: summary.Organization.Login,
			ProjectNumber:     summary.Organization.ProjectV2.Number,
			ProjectTitle:      summary.Organization.ProjectV2.Title,
			Item:              v,
		}

		line, err := json.Marshal(value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal json")
		}

		isLastNode := len(items.Node.Items.Nodes)-1 == i && hasNextPage == false
		if !isLastNode {
			line = append(line, byte('\n'))
		}
		lines = append(lines, line...)
	}

	return lines, nil
}
