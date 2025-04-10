package lib

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sisisin/gh-project-sync/lib/appcontext"
	"github.com/sisisin/gh-project-sync/lib/logger"
)

//go:embed get_viewer.graphql
var getViewerQuery string

//go:embed get_project_summary.graphql
var getProjectSummaryQuery string

//go:embed get_project_items.graphql
var getProjectItemsQuery string

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

type GitHubProjectSummary struct {
	RateLimit    RateLimit `json:"rateLimit"`
	Organization struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		Login     string
		ProjectV2 struct {
			Id     string `json:"id"`
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"projectV2"`
	} `json:"organization"`
}

func (c *GitHubClient) GetProjectSummary(ctx context.Context, org string, projectNumber int) (*GitHubProjectSummary, error) {
	variables := map[string]any{
		"org":           org,
		"projectNumber": projectNumber,
		"dryRun":        false,
	}

	res, err := request[GitHubProjectSummary](ctx, c.httpClient, getProjectSummaryQuery, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query GetProjectId")
	}

	if appcontext.GetVerbose(ctx) {
		logger.Info(ctx, "rateLimit", slog.Any("body", res.Data.RateLimit))
	}

	return &res.Data, nil
}

type RateLimit struct {
	Cost      int    `json:"cost"`
	Limit     int    `json:"limit"`
	NodeCount int    `json:"nodeCount"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"resetAt"`
	Used      int    `json:"used"`
}

type GitHubProjectItems struct {
	RateLimit RateLimit `json:"rateLimit"`
	Node      struct {
		Items struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			TotalCount int              `json:"totalCount"`
			Nodes      []map[string]any `json:"nodes"`
		} `json:"items"`
	} `json:"node"`
}

func (c *GitHubClient) GetProjectItems(ctx context.Context, id string, cursor string) (*GitHubProjectItems, error) {
	variables := map[string]any{
		"id":     id,
		"first":  100,
		"cursor": nil,
		"dryRun": appcontext.GetDryRun(ctx),
	}
	if cursor != "" {
		variables["after"] = cursor
	}

	res, err := request[GitHubProjectItems](ctx, c.httpClient, getProjectItemsQuery, variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query GetProjectItems")
	}
	return &res.Data, nil
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
