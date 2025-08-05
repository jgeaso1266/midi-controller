#!/bin/bash

set -e

OS=$(uname)

if [ "$OS" = "Linux" ]; then
    echo "Detected Linux OS. Installing libsoundio..."
    sudo apt-get install -y libasound2-dev
elif [ "$OS" = "Darwin" ]; then
    echo "Detected macOS (Darwin). Installing libsoundio..."
    brew install libsoundio
else
    echo "Unsupported OS: $OS"
    exit 1
fi
