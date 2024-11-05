package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

type ProjectItemsTable struct {
	client    *bigquery.Client
	datasetID string
	tableID   string
}

func New(ctx context.Context, datasetID, tableID string) (*ProjectItemsTable, error) {
	projectID := appcontext.GetProjectID(ctx)
	client, err := bigquery.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bigquery client")
	}

	client.Location = "US" // TODO: fix hard coded location
	return &ProjectItemsTable{
		client:    client,
		datasetID: datasetID,
		tableID:   tableID,
	}, nil
}

func (t *ProjectItemsTable) DeleteByProjectNumber(ctx context.Context, partitionTime time.Time, projectNumber int) error {
	query := fmt.Sprintf(
		"DELETE FROM `%s.%s.%s` "+
			"WHERE TIMESTAMP_TRUNC(_PARTITIONTIME, HOUR) = TIMESTAMP_TRUNC(\"%s\", HOUR) "+
			"AND project_number = %d",
		appcontext.GetProjectID(ctx), t.datasetID, t.tableID, partitionTime.Format(time.RFC3339), projectNumber)
	q := t.client.Query(query)
	job, err := q.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run query")
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to wait job")
	}
	if status.Done() {
		if err := status.Err(); err != nil {
			return errors.Wrap(err, "job failed")
		}
	}

	return nil
}

func (t *ProjectItemsTable) Load(ctx context.Context, loadTargetUri string, partitionTime time.Time) error {
	gcsRef := bigquery.NewGCSReference(loadTargetUri)
	gcsRef.SourceFormat = bigquery.JSON
	partition := partitionTime.Format("2006010215")
	loader := t.client.Dataset(t.datasetID).Table(fmt.Sprintf("%s$%s", t.tableID, partition)).LoaderFrom(gcsRef)
	loader.CreateDisposition = bigquery.CreateNever
	loader.WriteDisposition = bigquery.WriteAppend
	loader.TimePartitioning = &bigquery.TimePartitioning{
		Type: bigquery.HourPartitioningType,
	}

	job, err := loader.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run loader")
	}

	logger.Infof(ctx, "load start. job: %+v", job)

	status, err := job.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to wait job")
	}
	if status.Done() {
		if err := status.Err(); err != nil {
			return errors.Wrap(err, "job failed")
		}
	}

	logger.Infof(ctx, "load end. job: %+v", job)
	return nil
}
