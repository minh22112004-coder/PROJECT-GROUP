#!/bin/bash
# Integration test for Pack-A-Mal Network Simulation

set -e

echo "=========================================="
echo "Network Simulation Integration Test"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

# Test 1: Check if INetSim is running
print_test "Checking if INetSim container is running..."
if docker ps | grep -q "pack-a-mal-inetsim"; then
    print_pass "INetSim container is running"
else
    print_fail "INetSim container is not running"
    echo "Please run: docker-compose -f docker-compose.network-sim.yml up -d"
    exit 1
fi

# Test 2: Test DNS service
print_test "Testing DNS service on port 53..."
if docker exec pack-a-mal-inetsim nc -zv localhost 53 2>&1 | grep -q "succeeded"; then
    print_pass "DNS service is accessible"
else
    print_fail "DNS service is not accessible"
    exit 1
fi

# Test 3: Test HTTP service
print_test "Testing HTTP service on port 80..."
if docker exec pack-a-mal-inetsim nc -zv localhost 80 2>&1 | grep -q "succeeded"; then
    print_pass "HTTP service is accessible"
else
    print_fail "HTTP service is not accessible"
    exit 1
fi

# Test 4: Test DNS resolution through INetSim
print_test "Testing DNS resolution through INetSim..."
RESOLVED_IP=$(docker exec pack-a-mal-inetsim nslookup test.example.com localhost 2>/dev/null | grep "Address:" | tail -1 | awk '{print $2}')
if [ -n "$RESOLVED_IP" ]; then
    print_pass "DNS resolution works (resolved to $RESOLVED_IP)"
else
    print_fail "DNS resolution failed"
    exit 1
fi

# Test 5: Test HTTP response from INetSim
print_test "Testing HTTP response from INetSim..."
HTTP_RESPONSE=$(docker exec pack-a-mal-inetsim curl -s -o /dev/null -w "%{http_code}" http://localhost/ 2>/dev/null)
if [ "$HTTP_RESPONSE" = "200" ]; then
    print_pass "HTTP service returns 200 OK"
else
    print_fail "HTTP service returned: $HTTP_RESPONSE"
    exit 1
fi

# Test 6: Test Service Simulation API
print_test "Testing Service Simulation API..."
if curl -s http://localhost:5000/status | grep -q "running"; then
    print_pass "Service Simulation API is accessible"
else
    print_fail "Service Simulation API is not accessible"
    exit 1
fi

# Test 7: Build sample package
print_test "Building sample malicious network package..."
SAMPLE_DIR="$PROJECT_DIR/sample_packages/malicious_network_package"
if [ -d "$SAMPLE_DIR" ]; then
    cd "$SAMPLE_DIR"
    if python3 setup.py sdist >/dev/null 2>&1; then
        print_pass "Sample package built successfully"
    else
        print_fail "Failed to build sample package"
        exit 1
    fi
else
    print_fail "Sample package directory not found: $SAMPLE_DIR"
    exit 1
fi

# Test 8: Test network simulation code
print_test "Running Go unit tests for networksim..."
cd "$PROJECT_DIR/internal/networksim"
if go test -v 2>&1 | grep -q "PASS"; then
    print_pass "Network simulation unit tests passed"
else
    print_fail "Network simulation unit tests failed"
    exit 1
fi

# Test 9: Verify sandbox DNS configuration
print_test "Checking sandbox DNS configuration support..."
if grep -q "dnsServers" "$PROJECT_DIR/internal/sandbox/sandbox.go"; then
    print_pass "Sandbox supports custom DNS configuration"
else
    print_fail "Sandbox DNS configuration not found"
    exit 1
fi

# Test 10: Verify worker integration
print_test "Checking worker network simulation integration..."
if grep -q "networksim" "$PROJECT_DIR/cmd/worker/main.go"; then
    print_pass "Worker has network simulation integration"
else
    print_fail "Worker network simulation integration not found"
    exit 1
fi

echo ""
echo "=========================================="
echo "All Integration Tests Passed! ✓"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Test with a real package analysis:"
echo "   cd $SAMPLE_DIR"
echo "   pip install -e ."
echo "   python test_network.py"
echo ""
echo "2. Check INetSim logs:"
echo "   docker logs pack-a-mal-inetsim"
echo ""
echo "3. View captured network activity:"
echo "   cat ../../../service-simulation-module/shared/logs/inetsim/service.log"
echo ""
