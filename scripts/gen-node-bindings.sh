#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Path to this plugin, Note this must be an abolsute path on Windows (see #15)
PROTOC_GEN_TS_PATH=`which protoc-gen-ts`

# Path to the grpc_node_plugin
PROTOC_GEN_GRPC_PATH=`which grpc_tools_node_protoc_plugin`

# Directory to write generated code to (.js and .d.ts files)
OUT_DIR="$DIR/../proto_bindings/node/v1"

mkdir -p $OUT_DIR

protoc \
    --plugin="protoc-gen-ts=${PROTOC_GEN_TS_PATH}" \
    --plugin=protoc-gen-grpc=${PROTOC_GEN_GRPC_PATH} \
    --js_out="import_style=commonjs,binary:${OUT_DIR}" \
    --ts_out="service=grpc-node:${OUT_DIR}" \
    --grpc_out="${OUT_DIR}" \
    -I "$DIR/../rpc/v1/" \
    "$DIR/../rpc/v1/api.proto"