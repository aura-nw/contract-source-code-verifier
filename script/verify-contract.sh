#!/bin/bash
SOURCE_URL="$1"
BUILDER_IMAGE="$2"
EXPECTED_CHECKSUM="$3"
URL_OPTION="$4"
DOWNLOAD_FILE=download_contract.tar
DOWNLOAD_DIR=tmpdir/download_contract.tar

if [ "$URL_OPTION" == "0" ]; then
    wget --no-verbose -O "$DOWNLOAD_DIR" "$SOURCE_URL"
    SOURCE_CHECKSUM=$(sha256sum "$DOWNLOAD_DIR")
    cd tmpdir
    tar -x --strip-components 1 -f "$DOWNLOAD_FILE"
else 
    cd tmpdir
    git clone $SOURCE_URL
fi

RUSTFLAGS='-C link-arg=-s' cargo wasm
CARGO_CHECKSUM=$(sha256sum target/wasm32-unknown-unknown/release/*.wasm | awk '{print $1}')

# docker run --rm \
#     -v "$(pwd):/code" \
#     --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
#     --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
#     "$BUILDER_IMAGE"
# DOCKER_CHECKSUM=$(sha256sum target/wasm32-unknown-unknown/release/*.wasm | awk '{print $1}')

if [ "$CARGO_CHECKSUM" == "$EXPECTED_CHECKSUM" ]; then
    exit 0
# else if [ "$DOCKER_CHECKSUM" == "$EXPECTED_CHECKSUM" ]; then
#     exit 0
else 
    exit 1
fi