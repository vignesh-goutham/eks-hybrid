#!/usr/bin/env bash
# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/" && pwd -P)"

CALICO_VERSION="${1:-$(cat "${SCRIPT_ROOT}/VERSION")}"

ECR_ACCOUNT_ID="381492195191"

for region in "us-west-2" "us-west-1"; do
	registry="${ECR_ACCOUNT_ID}.dkr.ecr.${region}.amazonaws.com"

	for image in $(curl -s -L https://github.com/projectcalico/calico/releases/download/$CALICO_VERSION/metadata.yaml | yq  -r ".images[]" ); do
		repo_with_version=$image
		if [[ $image != quay* ]]; then
			repo_with_version="quay.io/$image"
		fi
		# inspecting the image, which returns the manifest/digests
		# will trigger the pull through cache if the image does not already exist in the repo.
		# using inspect instead of pull since we do not need the image locally
		docker buildx imagetools inspect "${registry}/${repo_with_version}"
	done
done