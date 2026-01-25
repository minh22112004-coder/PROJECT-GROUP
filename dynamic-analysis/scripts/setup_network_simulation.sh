#!/bin/bash

# Pack-A-Mal Network Simulation Setup Script
# This script helps set up Pack-A-Mal with INetSim network simulation

set -e

echo "=========================================="
echo "Pack-A-Mal Network Simulation Setup"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
print_info "Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi
print_info "✓ Docker found"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    print_warning "docker-compose command not found, trying 'docker compose'"
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi
print_info "✓ Docker Compose found"

# Determine project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
print_info "Project root: $PROJECT_ROOT"

# Create necessary directories
print_info "Creating necessary directories..."
mkdir -p "$PROJECT_ROOT/service-simulation-module/shared/logs/inetsim"
mkdir -p "$PROJECT_ROOT/service-simulation-module/shared/config/etc/inetsim"
print_info "✓ Directories created"

# Copy environment file if it doesn't exist
if [ ! -f "$SCRIPT_DIR/.env" ]; then
    print_info "Creating .env file from template..."
    cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"
    print_warning "Please edit .env file with your configuration"
else
    print_info "✓ .env file already exists"
fi

# Build and start INetSim services
print_info ""
print_info "Building and starting INetSim services..."
cd "$SCRIPT_DIR"

$DOCKER_COMPOSE -f docker-compose.network-sim.yml build
$DOCKER_COMPOSE -f docker-compose.network-sim.yml up -d inetsim service-simulation

# Wait for services to be healthy
print_info "Waiting for services to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker ps | grep -q "pack-a-mal-inetsim.*healthy"; then
        print_info "✓ INetSim is healthy"
        break
    fi
    attempt=$((attempt + 1))
    echo -n "."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    print_error "INetSim failed to become healthy"
    docker logs pack-a-mal-inetsim
    exit 1
fi

# Test INetSim connectivity
print_info ""
print_info "Testing INetSim connectivity..."

# Test DNS
if docker exec pack-a-mal-inetsim nc -zv localhost 53 2>&1 | grep -q "succeeded"; then
    print_info "✓ DNS service is accessible"
else
    print_warning "DNS service test failed"
fi

# Test HTTP
if docker exec pack-a-mal-inetsim nc -zv localhost 80 2>&1 | grep -q "succeeded"; then
    print_info "✓ HTTP service is accessible"
else
    print_warning "HTTP service test failed"
fi

# Test Service Simulation API
if curl -s http://localhost:5000/status | grep -q "running"; then
    print_info "✓ Service Simulation API is running"
else
    print_warning "Service Simulation API test failed"
fi

# Build sample package
print_info ""
print_info "Building sample malicious network package..."
SAMPLE_PKG_DIR="$SCRIPT_DIR/sample_packages/malicious_network_package"

if [ -d "$SAMPLE_PKG_DIR" ]; then
    cd "$SAMPLE_PKG_DIR"
    python3 setup.py sdist
    print_info "✓ Sample package built"
    print_info "  Package location: $SAMPLE_PKG_DIR/dist/"
else
    print_warning "Sample package directory not found"
fi

# Print summary
echo ""
echo "=========================================="
echo "Setup Complete!"
echo "=========================================="
echo ""
print_info "Services running:"
echo "  - INetSim DNS:  172.20.0.2:53"
echo "  - INetSim HTTP: 172.20.0.2:80"
echo "  - Sim API:      http://localhost:5000"
echo ""
print_info "Next steps:"
echo "  1. Configure Pack-A-Mal worker environment variables in .env"
echo "  2. Start the worker (if using Docker Compose):"
echo "     $DOCKER_COMPOSE -f docker-compose.network-sim.yml up pack-a-mal-worker"
echo ""
echo "  3. Or run analysis manually:"
echo "     export OSSF_NETWORK_SIMULATION_ENABLED=true"
echo "     export OSSF_INETSIM_DNS_ADDR=172.20.0.2:53"
echo "     ./scripts/run_analysis.sh --ecosystem pypi --package sample_packages/malicious_network_package"
echo ""
print_info "To view INetSim logs:"
echo "  docker logs -f pack-a-mal-inetsim"
echo ""
print_info "To stop services:"
echo "  $DOCKER_COMPOSE -f docker-compose.network-sim.yml down"
echo ""
print_info "For more information, see:"
echo "  - NETWORK_SIMULATION_GUIDE.md"
echo "  - sample_packages/malicious_network_package/README.md"
echo ""
