#!/bin/bash

# Detect the operating system
OS=$(uname -s)
ARCH=$(uname -m)

# Set Go version to 1.23.6
GO_VERSION="1.23.6"
GO_SHA256=""

calculate_sha256() {
    local file="$1"
    if command -v sha256sum &> /dev/null; then
        sha256sum "$file" | awk '{print $1}'
    elif command -v shasum &> /dev/null; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        echo "No SHA-256 checksum tool found (sha256sum or shasum)."
        exit 1
    fi
}

# Determine download URL and archive filename based on OS and ARCH
if [[ "$OS" == "Linux" ]]; then
    if [[ "$ARCH" == "x86_64" ]]; then
        GO_ARCHIVE="go$GO_VERSION.linux-amd64.tar.gz"
        GO_URL="https://go.dev/dl/$GO_ARCHIVE"
        GO_SHA256="9379441ea310de000f33a4dc767bd966e72ab2826270e038e78b2c53c2e7802d"
    elif [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
        GO_ARCHIVE="go$GO_VERSION.linux-arm64.tar.gz"
        GO_URL="https://go.dev/dl/$GO_ARCHIVE"
        GO_SHA256="561c780e8f4a8955d32bf72e46af0b5ee5e0debe1e4633df9a03781878219202"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi
elif [[ "$OS" == "Darwin" ]]; then
    if [[ "$ARCH" == "x86_64" ]]; then
        GO_ARCHIVE="go$GO_VERSION.darwin-amd64.tar.gz"
        GO_URL="https://go.dev/dl/$GO_ARCHIVE"
        GO_SHA256="782da50ce8ec5e98fac2cd3cdc6a1d7130d093294fc310038f651444232a3fb0"
    elif [[ "$ARCH" == "arm64" ]]; then
        GO_ARCHIVE="go$GO_VERSION.darwin-arm64.tar.gz"
        GO_URL="https://go.dev/dl/$GO_ARCHIVE"
        GO_SHA256="5cae2450a1708aeb0333237a155640d5562abaf195defebc4306054565536221"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi
else
    echo "Unsupported OS: $OS"
    exit 1
fi

# Parse command-line arguments
FORCE_CLEAN=false
CLEAR=false

# Process flags
for arg in "$@"; do
    case $arg in
        --clean)
            FORCE_CLEAN=true
            echo "Force clean enabled. Existing files will be removed."
            ;;
        --clear)
            CLEAR=true
            FORCE_CLEAN=true
            echo "Clear enabled. All new files will be removed (except files in ./tools/trond)."
            ;;
        *)
            ;;
    esac
done

# Check if Go is already installed on the system
if command -v go &> /dev/null; then
    echo "Go is already installed on the system: $(go version)"
    # SYSTEM_GO=true
    SYSTEM_GO=false
else
    SYSTEM_GO=false
fi

# If --clean or --clear is used, remove existing files
if [[ "$FORCE_CLEAN" == true ]]; then
    echo "Cleaning up existing Go files and binaries..."

    # Remove the downloaded archive if it exists
    if [[ -f "go/$GO_ARCHIVE" ]]; then
        rm -f "go/$GO_ARCHIVE"
        echo "Removed go/$GO_ARCHIVE"
    fi

    # Remove the extracted Go directory if it exists
    if [[ -d "go" ]]; then
        rm -rf go
        echo "Cleaned up Go directory"
    fi

    # Remove the trond binary if it exists
    if [[ -f "./tools/trond/trond" ]]; then
        rm -f "./tools/trond/trond"
        echo "Removed trond binary"
    fi

    # If clear is requested, stop here
    if [[ "$CLEAR" == true ]]; then
        exit 0
    fi
fi

# Create a directory for Go
if [[ ! -d "go" ]]; then
    mkdir -p go
    echo "Created go directory"
fi

# If Go is not installed, download and extract it
if [[ "$SYSTEM_GO" == false ]]; then
    # Check if the Go archive is already downloaded
    if [[ -f "go/$GO_ARCHIVE" ]]; then
        echo "go/$GO_ARCHIVE already exists. Skipping download."
    else
        echo "Downloading Go from $GO_URL..."
        curl -fL -o "go/$GO_ARCHIVE" "$GO_URL"
    fi

    # Verify the downloaded archive checksum
    echo "Verifying SHA-256 for go/$GO_ARCHIVE..."
    ACTUAL_SHA256="$(calculate_sha256 "go/$GO_ARCHIVE")"
    if [[ "$ACTUAL_SHA256" != "$GO_SHA256" ]]; then
        echo "SHA-256 mismatch for go/$GO_ARCHIVE"
        echo "Expected: $GO_SHA256"
        echo "Actual:   $ACTUAL_SHA256"
        rm -f "go/$GO_ARCHIVE"
        exit 1
    fi
    echo "SHA-256 verification passed."

    # Extract Golang to the go directory
    if [[ -d "go/bin" ]]; then
        echo "Go is already extracted. Skipping extraction."
    else
        echo "Extracting Go..."
        tar -xzf "go/$GO_ARCHIVE" -C go --strip-components=1
    fi

    # Set up Go binary path for the current directory
    GO_BIN="$(pwd)/go/bin"
    echo "Go binary path set to: $GO_BIN"
    export PATH="$GO_BIN:$PATH"

    # Set GOPATH to a separate workspace directory
    GOPATH="$(pwd)/gopath"
    echo "GOPATH set to: $GOPATH"
    export GOPATH
    mkdir -p "$GOPATH/src" "$GOPATH/pkg" "$GOPATH/bin"
else
    # Use the system Go binary location
    GO_BIN=$(dirname "$(command -v go)")
    echo "Using system Go binary: $GO_BIN"
fi

# Verify Go binary exists
if [[ ! -f "$GO_BIN/go" ]]; then
    echo "Error: Go binary not found at $GO_BIN/go"
    exit 1
fi

echo "Go setup successful. Version: $(go version)"

# Check for the main.go file in ./tools/trond/
MAIN_GO_PATH="./tools/trond/main.go"
if [[ ! -f "$MAIN_GO_PATH" ]]; then
    echo "Error: $MAIN_GO_PATH not found!"
    exit 1
fi

# Change to the tools/trond directory for building
echo "Changing to ./tools/trond/ directory..."
cd ./tools/trond || exit

# Build the Go program
echo "Building main.go..."
"$GO_BIN/go" build -ldflags="-s -w" -o trond main.go

# Check if the build was successful
if [[ -f "./trond" ]]; then
    echo "Build successful! The output binary is ./tools/trond/trond"
else
    echo "Build failed."
    exit 1
fi

# Move the binary to the original directory
echo "Moving the binary to the original directory..."
mv trond ../../

# Return to the original directory
echo "Returning to the original directory..."
cd - >/dev/null || exit

if [[ -f "./trond" ]]; then
    echo "Binary moved successfully! Run ./trond to execute."
else
    echo "Failed to move the binary."
    exit 1
fi

echo "Done!"
