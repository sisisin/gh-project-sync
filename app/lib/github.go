package lib

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

//go:embed get_viewer.graphql
var getViewerQuery string

//go:embed get_project_detail.graphql
var getProjectDetailQuery string

//go:embed get_project_id.graphql
var getProjectIDQuery string

const githubApiUrl = "https://api.github.com/graphql"

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.wrapped.RoundTrip(r)
}

type GitHubClient struct {
	httpClient *http.Client
}

func NewGithubClient(token string) *GitHubClient {
	httpClient := &http.Client{Transport: &authedTransport{token: token, wrapped: http.DefaultTransport}}
	return &GitHubClient{httpClient: httpClient}
}

func (c *GitHubClient) GetViewer(ctx context.Context) (map[string]any, error) {
	res, err := request[map[string]any](ctx, c.httpClient, getViewerQuery, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query GetViewer")
	}

	return res.Data, nil
}

func (c *GitHubClient) GetProjectID(ctx context.Context, org string, projectNumber int) (string, error) {
	variables := map[string]any{
		"org":           org,
		"projectNumber": projectNumber,
		"dryRun":        false,
	}

	res, err := request[map[string]any](ctx, c.httpClient, getProjectIDQuery, variables)
	if err != nil {
		return "", errors.Wrap(err, "failed to query GetProjectId")
	}

	if GetVerbose(ctx) {
		fmt.Println("rateLimit", res.Data["rateLimit"])
	}

	ret, ok := res.Data["organization"].(map[string]any)["projectV2"].(map[string]any)["id"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get project id: %v", res)
	}
	return ret, nil
}

func (c *GitHubClient) GetProjectDetail(ctx context.Context, id string) (map[string]any, error) {
	variables := map[string]any{
		"id":     id,
		"first":  40,
		"after":  nil,
		"dryRun": GetDryRun(ctx),
	}

	// todo: pagination
	// note: 1 req cost: 20, total count: 574
	res, err := request[map[string]any](ctx, c.httpClient, getProjectDetailQuery, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query GetProjectDetail")
	}
	return res.Data, nil
}

func request[T any](_ context.Context, httpClient *http.Client, query string, variables map[string]any) (*GraphqlResponse[T], error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}
	j, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal body")
	}

	res, err := httpClient.Post(githubApiUrl, "application/json", strings.NewReader(string(j)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	var respData GraphqlResponse[T]
	if err := json.Unmarshal(resBody, &respData); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}
	if len(respData.Errors) > 0 {
		errorsJson, _ := json.Marshal(respData.Errors)
		return nil, fmt.Errorf("graphql error occurred: %s", errorsJson)
	}

	return &respData, nil
}

type GraphqlResponse[T any] struct {
	Data   T              `json:"data"`
	Errors []GraphqlError `json:"errors"`
}

type GraphqlError struct {
	Locations []struct {
		Column int `json:"column"`
		Line   int `json:"line"`
	} `json:"locations"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
