package lib

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

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
