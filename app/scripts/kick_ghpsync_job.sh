#!/usr/bin/env bash

set -eu -o pipefail

gcloud run jobs execute github-project-sync --region=us-west1 --args="-verbose"
