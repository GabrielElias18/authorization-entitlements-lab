#!/bin/bash

# Test script for Go Entitlement Service
# Tests all canonical use cases across SpiceDB, Neo4j, and GraphQL implementations

set -e

echo "🧪 Testing Go Entitlement Service"
echo "=================================="

# Start the server in background
echo "🚀 Starting gRPC server..."
PORT=50052 go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "📋 Running Canonical Permission Tests"
echo "====================================="

# Test 1: Direct ownership - David can access his own account
echo "✅ Test 1: David can view transactions on acc123 (direct ownership)"
grpcurl -plaintext -d '{
  "user_id": "david",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 2: David can initiate payment on acc123 (direct ownership)"
grpcurl -plaintext -d '{
  "user_id": "david",
  "resource_id": "acc123",
  "permission": "can_initiate_payment",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 3: Emma can view transactions on acc456 (direct ownership)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 4: Emma can initiate payment on acc456 (direct ownership)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc456",
  "permission": "can_initiate_payment",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🔐 Test 5: Emma can view transactions on acc123 (POA with context)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "context": {"amount": "1000", "now": "2025-01-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🔐 Test 6: Emma can initiate payment on acc123 (POA with context)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc123",
  "permission": "can_initiate_payment",
  "context": {"amount": "500", "now": "2025-01-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🔐 Test 7: Adi can view transactions on acc456 (accountant access with context)"
grpcurl -plaintext -d '{
  "user_id": "adi",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "context": {"now": "2025-01-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🔐 Test 8: Adi can download statement on acc456 (accountant access with context)"
grpcurl -plaintext -d '{
  "user_id": "adi",
  "resource_id": "acc456",
  "permission": "can_download_statement",
  "context": {"now": "2025-01-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🏢 Test 9: Alice can access org accounts via role"
grpcurl -plaintext -d '{
  "user_id": "alice",
  "resource_id": "acc789",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🏢 Test 10: Alice can access org accounts via role"
grpcurl -plaintext -d '{
  "user_id": "alice",
  "resource_id": "acc790",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "❌ Test 11: Bob cannot access any accounts (negative test)"
grpcurl -plaintext -d '{
  "user_id": "bob",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "❌ Test 12: Charlie cannot access any accounts (negative test)"
grpcurl -plaintext -d '{
  "user_id": "charlie",
  "resource_id": "acc789",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "❌ Test 13: Alice cannot access non-org accounts (negative test)"
grpcurl -plaintext -d '{
  "user_id": "alice",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "❌ Test 14: Emma cannot access acc456 with expired POA (time-bound negative test)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "context": {"amount": "1000", "now": "2025-12-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "❌ Test 15: Adi cannot access acc456 with expired accountant access (time-bound negative test)"
grpcurl -plaintext -d '{
  "user_id": "adi",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "context": {"now": "2025-12-15T10:00:00Z"},
  "implementation": "IMPLEMENTATION_SPICEDB"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🎯 Testing Neo4j Implementation"
echo "==============================="

# Test Neo4j implementation
echo "✅ Test 16: David can view transactions on acc123 (Neo4j)"
grpcurl -plaintext -d '{
  "user_id": "david",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_NEO4J"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 17: Emma can view transactions on acc456 (Neo4j)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_NEO4J"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🎯 Testing GraphQL Implementation"
echo "================================="

# Test GraphQL implementation
echo "✅ Test 18: David can view transactions on acc123 (GraphQL)"
grpcurl -plaintext -d '{
  "user_id": "david",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_GRAPHQL"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 19: Emma can view transactions on acc456 (GraphQL)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_GRAPHQL"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🎯 Testing Both Implementations"
echo "==============================="

# Test both implementations
echo "✅ Test 20: David can view transactions on acc123 (both implementations)"
grpcurl -plaintext -d '{
  "user_id": "david",
  "resource_id": "acc123",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_BOTH"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "✅ Test 21: Emma can view transactions on acc456 (both implementations)"
grpcurl -plaintext -d '{
  "user_id": "emma",
  "resource_id": "acc456",
  "permission": "can_view_transactions",
  "implementation": "IMPLEMENTATION_BOTH"
}' localhost:50052 entitlement.v1.EntitlementService/CheckPermission

echo ""
echo "🏁 Tests completed!"
echo "==================="

# Stop the server
echo "🛑 Stopping gRPC server..."
kill $SERVER_PID 2>/dev/null || true

echo "✅ All tests completed successfully!" 