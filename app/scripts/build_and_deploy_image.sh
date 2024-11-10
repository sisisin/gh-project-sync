#!/usr/bin/env bash

set -ef -o pipefail

work_dir=$(cd "$(dirname "$0")" && pwd)
readonly work_dir

cd "$work_dir/.."

ts=$(TZ=JST-9 date "+%Y%m%d-%H%M%S")
image_id=us-west1-docker.pkg.dev/${PROJECT_ID}/github-project-sync/app:$ts

docker build --platform linux/amd64 -t "$image_id" .
gcloud auth configure-docker us-west1-docker.pkg.dev
docker push "$image_id"

echo "Done."
echo "$image_id"

sed -i '' "s|image: .*|image: $image_id|" job.yaml

gcloud run jobs replace job.yaml
