#!/usr/bin/env bash

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

git clone https://github.com/jenkins-x/jenkins-x-versions.git

BUILDER_VERSION=$(jx step get dependency-version --host=github.com --owner=jenkins-x --repo=jenkins-x-builders --short --dir jenkins-x-versions)

sed -i -e "s/\(FROM gcr\.io\/jenkinsxio\/builder-go-maven\:\).*/\1${BUILDER_VERSION}/" Dockerfile
