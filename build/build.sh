#!/usr/bin/env bash

set -euo pipefail
set -x

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

NAME="gflow"
IMAGE_NAME="olive-io/${NAME}"
GIT_COMMIT="$(git rev-parse --short HEAD)"
GIT_TAG="$(git describe --abbrev=0 --tags --always --match 'v*')"
GIT_VERSION="github.com/olive-io/gflow/pkg/version"
CGO_ENABLED="0"
BUILD_DATE="$(date +%s)"
LDFLAGS="-X ${GIT_VERSION}.GitCommit=${GIT_COMMIT} -X ${GIT_VERSION}.GitTag=${GIT_TAG} -X ${GIT_VERSION}.BuildDate=${BUILD_DATE}"
IMAGE_TAG="${GIT_TAG}-${GIT_COMMIT}"
REPO_ROOT="github.com/olive-io/gflow"

gflow::build::collect_proto_files() {
  mapfile -t TYPES_PROTO_FILES < <(find api/types -name '*.proto')
  mapfile -t RPC_PROTO_FILES < <(find api/rpc -name '*.proto')
  mapfile -t OPENAPI_PROTO_FILES < <(find api/rpc -name '*.proto')
}

gflow::build::vendor() {
  go mod vendor
}

gflow::build::test_coverage() {
  go test ./... -bench=. -coverage
}

gflow::build::lint() {
  golint -set_exit_status ./..
}

gflow::build::install() {
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install github.com/gogo/protobuf/protoc-gen-gofast@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.0
  go install github.com/srikrsna/protoc-gen-gotag@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
  go install github.com/envoyproxy/protoc-gen-validate@latest
}

gflow::build::apis() {
  gflow::build::collect_proto_files

  protoc --proto_path=./api \
    --proto_path=./third-party \
    --gofast_out=paths=source_relative:./api \
    "${TYPES_PROTO_FILES[@]}"

  protoc --proto_path=. \
    --proto_path=./api \
    --proto_path=./third-party \
    --gotag_out=paths=source_relative:. \
    "${TYPES_PROTO_FILES[@]}"

  protoc --proto_path=./api \
    --proto_path=./third-party \
    --go_out=paths=source_relative:./api \
    --go-grpc_out=paths=source_relative:./api \
    --validate_out=paths=source_relative,lang=go:./api \
    --grpc-gateway_out=paths=source_relative:./api \
    "${RPC_PROTO_FILES[@]}"

  protoc --proto_path=./api \
    --proto_path=./third-party \
    --openapi_out=fq_schema_naming=true,title="Gflow",description="Gflow OpenAPI3.0 Document",version=${GIT_TAG},default_response=true,paths=source_relative,:./server/docs \
    "${OPENAPI_PROTO_FILES[@]}"
}

gflow::build::docker() {
  :
}

gflow::build::vet() {
  go vet ./...
}

gflow::build::test() {
  vet
  go test -v ./...
}

gflow::build::clean() {
  rm -fr ./_output
}

gflow::build::build() {
  go build ./...
}

gflow::build::all() {
  gflow::build::build
}

gflow::build::usage() {
  cat <<'EOF'
Usage: build/build.sh <target>

Targets:
  all
  build
  vendor
  test-coverage
  lint
  install
  apis
  docker
  vet
  test
  clean
EOF
}

main() {
  local target="${1:-all}"

  case "$target" in
    all)
      gflow::build::all
      ;;
    build)
      gflow::build::build
      ;;
    vendor)
      gflow::build::vendor
      ;;
    test-coverage)
      gflow::build::test_coverage
      ;;
    lint)
      gflow::build::lint
      ;;
    install)
      gflow::build::install
      ;;
    apis)
      gflow::build::apis
      ;;
    docker)
      gflow::build::docker
      ;;
    vet)
      gflow::build::vet
      ;;
    test)
      gflow::build::test
      ;;
    clean)
      gflow::build::clean
      ;;
    -h|--help|help)
      gflow::build::usage
      ;;
    *)
      echo "Unknown target: ${target}" >&2
      gflow::build::usage
      exit 1
      ;;
  esac
}

main "$@"
