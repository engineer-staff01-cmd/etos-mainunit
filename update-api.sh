#!/bin/bash
set -eu

# install oapi-codegen
# See here https://github.com/deepmap/oapi-codegen
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest


for api in sync measure; do
  for type in client server types; do
    ~/go/bin/oapi-codegen -generate ${type} -package ${api}api api/${api}api.json > ${api}api/${api}_${type}.gen.go
  done
done