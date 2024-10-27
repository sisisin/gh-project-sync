package lib

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
