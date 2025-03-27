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
            sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose
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
            sleep 5
            sudo usermod -aG docker "$USER"
            newgrp docker
            if ! docker info &> /dev/null; then
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
