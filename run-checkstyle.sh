#!/bin/bash
set -eu # exit on any error

# ─── Configuration ───
CHECKSTYLE_VERSION="8.42"
CHECKSTYLE_JAR="checkstyle-${CHECKSTYLE_VERSION}-all.jar"
CHECKSTYLE_URL="https://github.com/checkstyle/checkstyle/releases/download/checkstyle-${CHECKSTYLE_VERSION}/${CHECKSTYLE_JAR}"

# SHA-256 digest of the official release JAR
# Obtain from: https://github.com/checkstyle/checkstyle/releases/tag/checkstyle-8.42
# or run: sha256sum checkstyle-8.42-all.jar
CHECKSTYLE_SHA256="4982ebeaa429fe41f3be2c3309a5c49d84c71ee1f78f967344b8bc82cf3101aa"

LIB_DIR="libs"
mkdir -p "$LIB_DIR"
CHECKSTYLE_PATH="${LIB_DIR}/${CHECKSTYLE_JAR}"

# ─── Download with integrity verification ───
verify_checksum() {
    local file="$1"
    local expected="$2"
    local actual
    actual=$(sha256sum "$file" | awk '{print $1}')
    if [ "$actual" != "$expected" ]; then
        echo "ERROR: Checksum mismatch for ${file}" >&2
        echo "  Expected: ${expected}" >&2
        echo "  Actual:   ${actual}" >&2
        rm -f "$file"
        exit 1
    fi
    echo "Checksum verified: ${file}"
}

if [ -f "$CHECKSTYLE_PATH" ]; then
    echo "Checkstyle JAR exists at ${CHECKSTYLE_PATH}, verifying integrity..."
    verify_checksum "$CHECKSTYLE_PATH" "$CHECKSTYLE_SHA256"
else
    echo "Downloading Checkstyle from ${CHECKSTYLE_URL}..."
    curl --fail --silent --show-error -L -o "$CHECKSTYLE_PATH" "$CHECKSTYLE_URL"
    verify_checksum "$CHECKSTYLE_PATH" "$CHECKSTYLE_SHA256"
    echo "Downloaded and verified successfully."
fi

# ─── .gitignore handling ───
GITIGNORE_FILE=".gitignore"
if [ -f "$GITIGNORE_FILE" ] && ! grep -q "^${LIB_DIR}/$" "$GITIGNORE_FILE"; then
    echo "${LIB_DIR}/" >> "$GITIGNORE_FILE"
fi

# ─── Find and check Java files ───
CHECKSTYLE_CONFIG="./conf/checkstyle/checkStyleAll.xml"

# Use find with -print0 / xargs -0 to handle filenames with spaces
JAVA_COUNT=$(find . -name "*.java" | wc -l)
if [ "$JAVA_COUNT" -eq 0 ]; then
    echo "No Java files found."
    exit 0
fi

echo "Found ${JAVA_COUNT} Java files, running Checkstyle..."

# Capture exit code correctly (not after echo)
find . -name "*.java" -print0 | xargs -0 java -jar "$CHECKSTYLE_PATH" -c "$CHECKSTYLE_CONFIG"
STATUS=$?

echo "Checkstyle finished with exit code: ${STATUS}"
exit $STATUS
