#!/bin/bash

# Check and install Docker if not present
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "Docker not found. Installing Docker..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # MacOS
            brew install --cask docker
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            sudo sh get-docker.sh
            sudo systemctl start docker
            sudo systemctl enable docker
            sudo usermod -aG docker "$USER"
            echo "Added current user to the docker group. You may need to log out and back in for this to take effect."
        else
            echo "Unsupported operating system"
            exit 1
        fi
    fi
}

# Check and install Docker Compose if not present
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        echo "Docker Compose not found. Installing Docker Compose..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # MacOS - Docker Compose comes with Docker Desktop
            echo "Please install Docker Desktop which includes Docker Compose"
            exit 1
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            # Linux
            compose_os="$(uname -s | tr '[:upper:]' '[:lower:]')"
            compose_arch="$(uname -m)"
            compose_binary="docker-compose-${compose_os}-${compose_arch}"
            compose_version="$(curl -fsSLI https://github.com/docker/compose/releases/latest \
                | awk -F'/' '/^location:/ {gsub(/\r/,"",$NF); print $NF}')"

            if [[ -z "$compose_version" ]]; then
                echo "Failed to determine latest Docker Compose version"
                exit 1
            fi

            tmp_dir="$(mktemp -d)"
            compose_url="https://github.com/docker/compose/releases/download/${compose_version}/${compose_binary}"
            checksums_url="https://github.com/docker/compose/releases/download/${compose_version}/checksums.txt"

            if ! curl -fsSL "$compose_url" -o "${tmp_dir}/${compose_binary}"; then
                rm -rf "$tmp_dir"
                echo "Failed to download Docker Compose binary: ${compose_url}"
                exit 1
            fi

            if ! curl -fsSL "$checksums_url" -o "${tmp_dir}/checksums.txt"; then
                rm -rf "$tmp_dir"
                echo "Failed to download Docker Compose checksums: ${checksums_url}"
                exit 1
            fi

            if ! grep -E "[[:space:]]\\*${compose_binary}$" "${tmp_dir}/checksums.txt" \
                | (cd "$tmp_dir" && sha256sum -c -); then
                rm -rf "$tmp_dir"
                echo "Docker Compose checksum verification failed"
                exit 1
            fi

            sudo install -m 0755 "${tmp_dir}/${compose_binary}" /usr/local/bin/docker-compose
            rm -rf "$tmp_dir"
        fi
    fi
}

# Check if Docker daemon is running
check_docker_running() {
    echo "Checking Docker daemon status..."
    if ! docker info &> /dev/null; then
        echo "Docker daemon is not running"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo "Please start Docker Desktop manually"
            exit 1
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            echo "Starting Docker daemon..."
            sudo systemctl start docker
            sleep 10
            # Only add non-root users to the docker group
            if [ "$EUID" -ne 0 ]; then
                sudo usermod -aG docker "$USER"
                echo "Added current user to the docker group."
                echo "You may need to log out and back in for this to take effect, or try 'newgrp docker' to reload your group membership."
                echo "Skipping Docker network configuration check, as you are not root"
                exit 0
            fi
            if ! sudo docker info &> /dev/null; then
                echo "Failed to start Docker daemon"
                exit 1
            fi
        fi
    fi
    echo "Docker daemon is running"
}

# Check Docker network configuration
check_docker_network() {
    echo "Checking Docker network configuration..."

    # Test Docker network creation
    if ! docker network create test-network &> /dev/null; then
        echo "Failed to create test network"
        exit 1
    fi

    # Test DNS resolution and network connectivity with AWS domains
    if ! docker run --rm --network test-network alpine nslookup amazonaws.com &> /dev/null; then
        echo "Docker network configuration issue: Cannot resolve AWS domains"
        docker network rm test-network &> /dev/null
        exit 1
    fi

    # Additional AWS regional endpoint test
    if ! docker run --rm --network test-network alpine nslookup s3.amazonaws.com &> /dev/null; then
        echo "Docker network configuration issue: Cannot resolve AWS S3 endpoint"
        docker network rm test-network &> /dev/null
        exit 1
    fi

    # Clean up test network
    docker network rm test-network &> /dev/null
    echo "Docker network configuration is working properly with AWS domains"
}

# Run checks
check_docker
check_docker_compose
check_docker_running
check_docker_network
