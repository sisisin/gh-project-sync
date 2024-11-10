#!/usr/bin/env bash

set -eu -o pipefail

script_dir=$(cd "$(dirname "$0")" && pwd)
readonly script_dir

region=$(pulumi --cwd="${script_dir}/../../infra" stack output region)
run_name=$(pulumi --cwd="${script_dir}/../../infra" stack output runName)
gcloud run jobs execute "${run_name}" --region="${region}" --args="/app/ghpsync" --args="-verbose"
