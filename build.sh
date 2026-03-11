#!/bin/bash
set -eu

export GOOS="linux"
export CGO_ENABLED=1

printUsage() {
    echo "build.sh [release|debug] [A9E|G3L]"
}

MODEL=""
if [ $# -ne 2 ] && [ "$2" != "A9E" ] && [ "$2" != "G3L" ]; then
    printUsage
    exit 1
else
    if [ "A9E" = $2 ]; then
        export GOARM=""
        export GOARCH="arm64"
        export CC="aarch64-linux-gnu-gcc"
        MODEL="A9E"
    elif [ "G3L" = $2 ]; then
        export GOARM="7"
        export GOARCH="arm"
        export CC="arm-linux-gnueabihf-gcc"
        MODEL="G3L"
    fi
fi

# 前回のMODEL値を確認し、異なる場合のみgo cleanを実行
BUILD_MODEL_FILE=".build_model"
PREVIOUS_MODEL=""
if [ -f "$BUILD_MODEL_FILE" ]; then
    PREVIOUS_MODEL=$(cat "$BUILD_MODEL_FILE" 2>/dev/null || echo "")
fi

if [ "$MODEL" != "$PREVIOUS_MODEL" ]; then
    echo "Model changed from '${PREVIOUS_MODEL:-none}' to '$MODEL', cleaning cache..."
    go clean -modcache -cache
    echo "$MODEL" > "$BUILD_MODEL_FILE"
else
    echo "Model unchanged ('$MODEL'), skipping cache clean"
fi

# Version, Revision 設定
REV=$(git rev-parse --short HEAD 2>/dev/null)
VER=$(git tag -l --contains $REV 2>/dev/null)
if [ -z $VER ]; then
    VER=$(date +"%Y/%m/%d_%H:%M:%S")
fi

# ビルド
RELEASE_FLAGS=""
if [ ! -z ${1+x} ] && [ "debug" = $1 ]; then
    echo "Build for debug"
    ENV="Staging"
else
    echo "Build for release"
    ENV="Production"
    # Omit the debug information
    RELEASE_FLAGS="-w -s"
fi

go build -trimpath -ldflags "-X etos-mainunit/app.VERSION=$VER -X etos-mainunit/app.REVISION=$REV -X etos-mainunit/app.ENV=$ENV -X etos-mainunit/app.MODEL=$MODEL $RELEASE_FLAGS"
if [ $? -gt 0 ]; then
    # エラー処理
    echo "=========================================="
    echo "===            BUILD ERROR             ==="
    echo "=========================================="
    exit
fi
echo "VER = $VER"
echo "REV = $REV"
echo "ENV = $ENV"
md5sum etos-mainunit
