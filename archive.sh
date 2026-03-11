#!/bin/sh
set -eu

archive() {
    local BINARY_FILENAME=$1
    local CHECKSUM_FILENAME=$2
    local ARCHIVE_FILENAME=$3

    CHECKSUM=$(md5sum ${BINARY_FILENAME} | cut -d ' ' -f 1 | sed -z 's/\n//g')
    echo -n ${CHECKSUM} >${CHECKSUM_FILENAME}
    zip ${ARCHIVE_FILENAME} ${CHECKSUM_FILENAME} ${BINARY_FILENAME}
}

archive "etos-mainunit" "md5sum" "archive.zip"
