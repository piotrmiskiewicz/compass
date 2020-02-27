#!/usr/bin/env bash
print() {
  echo "> $1"
}
build_binary() {
  name=$1
  path_to_main=$2
  env GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $name $path_to_main
}
print "Build broker binary"
build_binary kyma-env-broker ./cmd/broker/
print "Build docker image"
docker build -t kyma-env-broker -f Dockerfile.local .
print "Tag docker image"
docker tag kyma-env-broker:latest eu.gcr.io/sap-se-cx-gopher/kyma-env-broker-test:latest
print "Push docker image"
docker push eu.gcr.io/sap-se-cx-gopher/kyma-env-broker-test:latest
print "Remove binary"
rm -rf kyma-env-broker
