package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var projectID *string

func init() {
	projectID = flag.String("project-id", "", "The Google Cloud project id to use")
	flag.Parse()

	if *projectID == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	fmt.Println("start init")

	repoRoot := getRepoRoot()
	{
		pulumiDevYamlPath := path.Join(repoRoot, "infra", "Pulumi.dev.yaml")
		jobYamlContent := string(must1(os.ReadFile(pulumiDevYamlPath)))
		re := regexp.MustCompile(`gcp:project: \S+`)
		match := re.FindIndex([]byte(jobYamlContent))
		if match == nil {
			panic("Google Cloud Project setup failed. Failed to replace project id in Pulumi.dev.yaml")
		}
		replaced := re.ReplaceAllString(jobYamlContent, "gcp:project: "+*projectID)
		must0(os.WriteFile(pulumiDevYamlPath, []byte(replaced), 0644))
	}

	fmt.Println("end init")
}
func getRepoRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output := must1(cmd.Output())
	return strings.TrimSpace(string(output))
}

func must0(err error) {
	if err != nil {
		panic(err)
	}
}
func must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
