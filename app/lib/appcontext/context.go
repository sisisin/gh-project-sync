package appcontext

import "context"

type dryRunKeyType string

var dryRunKey dryRunKeyType = "dryRun"

func WithDryRun(ctx context.Context, dryRun bool) context.Context {
	return context.WithValue(ctx, dryRunKey, dryRun)
}

func GetDryRun(ctx context.Context) bool {
	dryRun, _ := ctx.Value(dryRunKey).(bool)
	return dryRun
}

type verboseKeyType string

var verboseKey verboseKeyType = "verbose"

func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseKey, verbose)
}
func GetVerbose(ctx context.Context) bool {
	verbose, _ := ctx.Value(verboseKey).(bool)
	return verbose
}

type projectIDKeyType string

const projectIDKey projectIDKeyType = "projectID"

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}
func GetProjectID(ctx context.Context) string {
	projectID, ok := ctx.Value(projectIDKey).(string)
	if !ok {
		panic("projectID not found in context")
	}
	return projectID
}

type TraceIDKeyType string

const TraceIDKey TraceIDKeyType = "traceID"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
func GetTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(TraceIDKey).(string)
	return traceID
}
