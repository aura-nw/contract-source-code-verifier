#!/bin/bash
SOURCE_URL="$1"
COMMIT="$2"
EXPECTED_CHECKSUM="$3"
DIR="$4"
CONTRACT_FOLDER="$5"

cd $DIR
git clone $SOURCE_URL
cd $CONTRACT_FOLDER
git checkout $COMMIT

# if [ "$COMPILER_IMAGE" == "" ]; then
    # RUSTFLAGS='-C link-arg=-s' cargo wasm
    # CARGO_CHECKSUM=$(sha256sum target/wasm32-unknown-unknown/release/*.wasm | awk '{print $1}')
    # echo $CARGO_CHECKSUM
# else
    docker run --rm \
        -v "$(pwd):/code" \
        --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
        --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
        $COMMIT
    CARGO_CHECKSUM=$(sha256sum artifacts/*.wasm | awk '{print $1}')
    echo $CARGO_CHECKSUM
# fi

if [ "$CARGO_CHECKSUM" == "$EXPECTED_CHECKSUM" ]; then
    cargo schema
    exit 0
else 
    exit 1
fi