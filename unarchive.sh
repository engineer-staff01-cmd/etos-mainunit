#!/bin/sh
set -eu

unarchive() {
    local BINARY_FILENAME=$1
    local OUTPUT_FOLDER=$2

    unzip ${BINARY_FILENAME} -d ${OUTPUT_FOLDER}
}

unarchive "archive.zip" archive
