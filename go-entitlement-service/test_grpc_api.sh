#!/bin/bash

# Comprehensive gRPC API Test Suite for Go Entitlement Service
# Tests SpiceDB vs Neo4j implementations with canonical test cases

set -e

GRPC_SERVER="localhost:50052"
SERVICE="entitlement.v1.EntitlementService"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Go Entitlement Service gRPC Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print test results
print_result() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    local backend="$4"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✓ PASS${NC} [$backend] $test_name (expected: $expected, got: $actual)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC} [$backend] $test_name (expected: $expected, got: $actual)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Function to test permission
test_permission() {
    local test_name="$1"
    local actor="$2"
    local resource="$3"
    local permission="$4"
    local expected="$5"
    local implementation="$6"
    local context_extra="$7"

    local context_json='{"implementation": "'$implementation'"'
    if [ -n "$context_extra" ]; then
        context_json="${context_json}, ${context_extra}"
    fi
    context_json="${context_json}}"

    local request='{
        "actor": "'$actor'",
        "resource": "'$resource'",
        "permission": "'$permission'",
        "context": '$context_json'
    }'

    local response
    response=$(grpcurl -plaintext -d "$request" $GRPC_SERVER $SERVICE/CheckPermission 2>/dev/null || echo "ERROR")

    if [ "$response" = "ERROR" ]; then
        print_result "$test_name" "$expected" "ERROR" "$implementation"
        return
    fi

    local has_permission
    has_permission=$(echo "$response" | jq -r '.hasPermission // false')

    print_result "$test_name" "$expected" "$has_permission" "$implementation"
}

# Check if server is running
echo -e "${YELLOW}Checking if gRPC server is running on $GRPC_SERVER...${NC}"
if ! grpcurl -plaintext $GRPC_SERVER $SERVICE/Health >/dev/null 2>&1; then
    echo -e "${RED}ERROR: gRPC server is not running on $GRPC_SERVER${NC}"
    echo "Please start the server with: go run cmd/server/main.go"
    exit 1
fi
echo -e "${GREEN}✓ gRPC server is running${NC}"
echo ""

# Test Health Check
echo -e "${BLUE}=== Health Check Tests ===${NC}"
health_response=$(grpcurl -plaintext -d '{"implementation": "IMPLEMENTATION_SPICEDB"}' $GRPC_SERVER $SERVICE/Health 2>/dev/null)
if echo "$health_response" | jq -e '.healthy == true' >/dev/null; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
fi
echo ""

# Core Permission Tests for Both Backends
run_permission_tests() {
    local implementation="$1"
    local backend_name="$2"

    echo -e "${BLUE}=== $backend_name Permission Tests ===${NC}"

    # David's direct ownership tests
    test_permission "David can download his own statement" "david" "acc123" "can_download_statement" "true" "$implementation"
    test_permission "David can initiate payment on his account" "david" "acc123" "can_initiate_payment" "true" "$implementation" '"amount": "1000"'
    test_permission "David can view transactions on his account" "david" "acc123" "can_view_transactions" "true" "$implementation"

    # Emma's direct ownership tests
    test_permission "Emma can download her own statement" "emma" "acc456" "can_download_statement" "true" "$implementation"
    test_permission "Emma can initiate payment on her account" "emma" "acc456" "can_initiate_payment" "true" "$implementation" '"amount": "1000"'

    # POA delegation tests - Emma has POA on David's account
    test_permission "Emma can download statement via POA" "emma" "acc123" "can_download_statement" "true" "$implementation"
    test_permission "Emma can initiate small payment via POA" "emma" "acc123" "can_initiate_payment" "true" "$implementation" '"amount": "500"'
    test_permission "Emma cannot initiate large payment via POA" "emma" "acc123" "can_initiate_payment" "false" "$implementation" '"amount": "6000"'

    # Adi's accountant access tests
    test_permission "Adi can download statement as accountant" "adi" "acc456" "can_download_statement" "true" "$implementation"
    test_permission "Adi can view transactions as accountant" "adi" "acc456" "can_view_transactions" "true" "$implementation"
    test_permission "Adi cannot initiate payments as accountant" "adi" "acc456" "can_initiate_payment" "false" "$implementation"

    # Role-based access tests - Alice has finance_ops role
    test_permission "Alice can access org account via role" "alice" "acc789" "can_access" "true" "$implementation"
    test_permission "Alice can download org statement via role" "alice" "acc789" "can_download_statement" "true" "$implementation"

    # Negative tests - access denied
    test_permission "Charlie cannot access David's account" "charlie" "acc123" "can_download_statement" "false" "$implementation"
    test_permission "Bob cannot access Emma's account" "bob" "acc456" "can_initiate_payment" "false" "$implementation"
    test_permission "Unknown user cannot access any account" "unknown_user" "acc123" "can_download_statement" "false" "$implementation"
    test_permission "David cannot access unknown account" "david" "unknown_acc" "can_download_statement" "false" "$implementation"

    echo ""
}

# Run tests for each backend
run_permission_tests "spicedb" "SpiceDB"
run_permission_tests "neo4j" "Neo4j GraphDB"
run_permission_tests "graphql" "GraphQL PostgreSQL"

# Comparison tests using "both" implementation
echo -e "${BLUE}=== Backend Comparison Tests ===${NC}"
test_permission "Comparison: David's ownership" "david" "acc123" "can_download_statement" "true" "both"
test_permission "Comparison: Emma's POA delegation" "emma" "acc123" "can_initiate_payment" "true" "both" '"amount": "500"'
test_permission "Comparison: Access denied case" "charlie" "acc123" "can_download_statement" "false" "both"
echo ""

# Bulk Permission Test
echo -e "${BLUE}=== Bulk Permission Test ===${NC}"
bulk_request='{
    "requests": [
        {
            "actor": "david",
            "resource": "acc123",
            "permission": "can_download_statement",
            "context": {"implementation": "spicedb"}
        },
        {
            "actor": "emma",
            "resource": "acc456",
            "permission": "can_download_statement",
            "context": {"implementation": "neo4j"}
        },
        {
            "actor": "charlie",
            "resource": "acc123",
            "permission": "can_download_statement",
            "context": {"implementation": "both"}
        }
    ],
    "maxConcurrency": 3
}'

bulk_response=$(grpcurl -plaintext -d "$bulk_request" $GRPC_SERVER $SERVICE/CheckBulkPermissions 2>/dev/null || echo "ERROR")
if [ "$bulk_response" != "ERROR" ]; then
    success_count=$(echo "$bulk_response" | jq -r '.successCount // 0')
    total_responses=$(echo "$bulk_response" | jq -r '.responses | length')
    echo -e "${GREEN}✓ Bulk request processed: $success_count/$total_responses successful${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ Bulk request failed${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# Simple Benchmark Test
echo -e "${BLUE}=== Simple Benchmark Test ===${NC}"
benchmark_request='{
    "testCases": [
        {
            "name": "David Ownership Test",
            "actor": "david",
            "resource": "acc123",
            "permission": "can_download_statement",
            "context": {"implementation": "spicedb"},
            "expectedResult": true
        }
    ],
    "iterations": 10,
    "concurrency": 2,
    "implementation": "IMPLEMENTATION_SPICEDB"
}'

benchmark_response=$(grpcurl -plaintext -d "$benchmark_request" $GRPC_SERVER $SERVICE/Benchmark 2>/dev/null || echo "ERROR")
if [ "$benchmark_response" != "ERROR" ]; then
    avg_time=$(echo "$benchmark_response" | jq -r '.results[0].avgResponseTimeMs // 0')
    throughput=$(echo "$benchmark_response" | jq -r '.summary.throughputRps // 0')
    echo -e "${GREEN}✓ Benchmark completed: ${avg_time}ms avg response time, ${throughput} req/sec throughput${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}✗ Benchmark failed${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
echo ""

# Final Results
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Results Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi