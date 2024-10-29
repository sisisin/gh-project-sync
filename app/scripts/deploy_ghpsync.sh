#!/usr/bin/env bash

set -eu -o pipefail

work_dir=$(cd "$(dirname "$0")" && pwd)
readonly work_dir

cd "$work_dir/.."

image_id="$1"

if [ -z "$image_id" ]; then
    echo "Usage: $0 <image_id>"
    exit 1
fi

sed -i '' "s|image: .*|image: $image_id|" job.yaml

gcloud run jobs replace job.yaml
