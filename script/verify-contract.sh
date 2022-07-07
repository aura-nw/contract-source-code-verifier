#!/bin/bash
SOURCE_URL="$1"
COMMIT="$2"
EXPECTED_CHECKSUM="$3"
DIR="$4"
CONTRACT_FOLDER="$5"
COMPILER_IMAGE="$6"
WASM_FILE="$7"
CONTRACT_DIR="$8"
TEMP_DIR="$9"
CODE_ID="${10}"

cd $DIR
git clone $SOURCE_URL
cd $CONTRACT_FOLDER
git checkout $COMMIT
rm -rf artifacts

docker run --rm \
    -v "$(pwd):/code" \
    --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
    --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
    $COMPILER_IMAGE
CARGO_CHECKSUM=$(sha256sum artifacts/$WASM_FILE | awk '{print $1}')
echo $CARGO_CHECKSUM

if [ "$CARGO_CHECKSUM" == "$EXPECTED_CHECKSUM" ]; then
    pwd
    if [ "$CONTRACT_DIR" != "" ]; then 
        cd $CONTRACT_DIR
    fi
    cargo clean
    cargo schema
    while [ "$(basename $PWD)" != "$TEMP_DIR" ]; do cd ..; done
    pwd
    zip -r code_id_$CODE_ID.zip $CONTRACT_FOLDER
    exit 0
else 
    exit 1
fi