#!/usr/bin/env bash

set -ef -o pipefail

script_dir=$(cd "$(dirname "$0")" && pwd)
readonly script_dir

cd "$script_dir/.."

ts=$(TZ=JST-9 date "+%Y%m%d-%H%M%S")
image_name=$(pulumi --cwd="${script_dir}/../../infra" stack output imageName)
registry_domain=$(pulumi --cwd="${script_dir}/../../infra" stack output registryDomain)
image_id=${image_name}:$ts

docker build --platform linux/amd64 -t "$image_id" .
gcloud auth configure-docker "${registry_domain}"
docker push "$image_id"

echo "Done."
echo "$image_id"

sed -i '' "s|image: .*|image: $image_id|" job.yaml

gcloud run jobs replace job.yaml
