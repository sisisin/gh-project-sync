package lib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
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

	fmt.Println("rateLimit:", projectDetail["rateLimit"])

	j, err := json.Marshal(projectDetail)
	if err != nil {
		return errors.Wrap(err, "failed to marshal projectDetail")
	}

	fmt.Println(string(j))
	return nil
}
