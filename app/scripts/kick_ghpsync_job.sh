#!/usr/bin/env bash

set -eu -o pipefail

script_dir=$(cd "$(dirname "$0")" && pwd)
readonly script_dir

region=$(pulumi --cwd="${script_dir}/../../infra" stack output region)
gcloud run jobs execute github-project-sync --region="${region}" --args="/app/ghpsync" --args="-verbose"
