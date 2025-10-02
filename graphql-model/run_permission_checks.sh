#!/bin/bash
set -e

# Database connection
PGHOST=localhost
PGPORT=5434
PGUSER=postgres
PGDATABASE=entitlements
PGPASSWORD=postgres
export PGPASSWORD
GRAPHQL_URL="http://localhost:4000/"

# Extract user IDs (using simple IDs without prefixes)
DAVID_ID="david"
EMMA_ID="emma"
ADI_ID="adi"
ALICE_ID="alice"

function check_permission() {
  local user_id="$1"
  local query="$2"
  local expected="$3"
  local description="$4"

  if [[ -z "$user_id" ]]; then
    echo "SKIP: $description (user_id is empty)"
    return
  fi

  raw_response=$(curl -s -X POST "$GRAPHQL_URL" \
    -H "Content-Type: application/json" \
    -H "x-user-id: $user_id" \
    --data-raw "$(jq -nc --arg q "$query" '{query: $q}')")

  echo "DEBUG: $description"
  echo "User ID: $user_id"
  echo "Query: $query"
  echo "Raw response: $raw_response"

  # Check for GraphQL errors
  if echo "$raw_response" | jq -e '.errors' >/dev/null; then
    echo "FAIL: $description (GraphQL error: $(echo "$raw_response" | jq -c '.errors'))"
    return
  fi

  result=$(echo "$raw_response" | jq -r '.data | to_entries[0].value')

  if [[ "$result" == "$expected" ]]; then
    echo "PASS: $description"
  else
    echo "FAIL: $description (got $result, expected $expected)"
  fi
}

###############################################
# Emma POA (limit) edge cases on acc123
# - POA grants emma permission to initiate payments on acc123
# - Caveat: maxAmount = 5000
# - Test just below, at, and just above the limit
###############################################
check_permission "$EMMA_ID" "query { canInitiatePayment(accountId: \"acc123\", amount: 4999) }" "true" "Emma can initiate payment on acc123 just below limit (4999)"
check_permission "$EMMA_ID" "query { canInitiatePayment(accountId: \"acc123\", amount: 5000) }" "true" "Emma can initiate payment on acc123 at limit (5000)"
check_permission "$EMMA_ID" "query { canInitiatePayment(accountId: \"acc123\", amount: 5001) }" "false" "Emma cannot initiate payment on acc123 just above limit (5001)"

###############################################
# Adi POA (time) edge cases on acc456
# - POA grants adi permission to initiate payments on acc456
# - Caveat: validFrom = 2024-01-01, validTo = 2024-03-01 (EXPIRED)
# - Test time-bound access (should fail as POA is expired)
###############################################
echo "(NOTE: Adi's POA expired on 2024-03-01, so all time-bound checks should fail)"
check_permission "$ADI_ID" "query { canInitiatePayment(accountId: \"acc456\", amount: 100) }" "false" "Adi cannot initiate payment on acc456 (expired POA - ended 2024-03-01)"

###############################################
# David (owner) and Alice (org/role) checks
# - David is the direct owner of acc123
# - Alice is a member of finance_ops role in org:abc, which owns acc789
###############################################
check_permission "$DAVID_ID" "query { canInitiatePayment(accountId: \"acc123\", amount: 1000) }" "true" "David can initiate payment on acc123 (owner)"
check_permission "$ALICE_ID" "query { canInitiatePayment(accountId: \"acc789\", amount: 100) }" "true" "Alice (finance_ops) can initiate payment on org:abc account acc789"

###############################################
# Negative: Bob (not in canonical data)
# - Bob should not be able to initiate payment on acc123 (not owner, not POA)
###############################################
BOB_ID="bob"
check_permission "$BOB_ID" "query { canInitiatePayment(accountId: \"acc123\", amount: 100) }" "false" "Bob cannot initiate payment on acc123 (not owner, not POA)"

###############################################
# Statement Download Tests
# - Test various users' ability to download statements for different accounts
###############################################
echo ""
echo "=== STATEMENT DOWNLOAD TESTS ==="

# David (owner) can download statement for acc123
check_permission "$DAVID_ID" "query { canDownloadStatement(accountId: \"acc123\") }" "true" "David can download statement for acc123 (owner)"

# Emma (POA) can download statement for acc123
check_permission "$EMMA_ID" "query { canDownloadStatement(accountId: \"acc123\") }" "true" "Emma can download statement for acc123 (POA)"

# Emma (owner) can download statement for acc456
check_permission "$EMMA_ID" "query { canDownloadStatement(accountId: \"acc456\") }" "true" "Emma can download statement for acc456 (owner)"

# Alice (org role) can download statement for acc789
check_permission "$ALICE_ID" "query { canDownloadStatement(accountId: \"acc789\") }" "true" "Alice can download statement for acc789 (org role)"

# Alice (org role) can download statement for acc790
check_permission "$ALICE_ID" "query { canDownloadStatement(accountId: \"acc790\") }" "true" "Alice can download statement for acc790 (org role)"

# Adi (expired POA) cannot download statement for acc456
check_permission "$ADI_ID" "query { canDownloadStatement(accountId: \"acc456\") }" "false" "Adi cannot download statement for acc456 (expired POA)"

# David cannot download statement for acc456 (no access)
check_permission "$DAVID_ID" "query { canDownloadStatement(accountId: \"acc456\") }" "false" "David cannot download statement for acc456 (no access)"

# Emma cannot download statement for acc789 (no org membership)
check_permission "$EMMA_ID" "query { canDownloadStatement(accountId: \"acc789\") }" "false" "Emma cannot download statement for acc789 (no org membership)"

# Test with unknown user (should be denied)
check_permission "user-unknown" "query { canDownloadStatement(accountId: \"acc123\") }" "false" "Unknown user cannot download statement for acc123"

# Test with empty user ID (should be denied)
check_permission "" "query { canDownloadStatement(accountId: \"acc123\") }" "false" "Empty user ID cannot download statement for acc123"

# Add more checks for org/role traversal, edge cases, etc. as needed 