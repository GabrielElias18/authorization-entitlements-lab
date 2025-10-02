#!/bin/bash
set -e

pass_count=0
fail_count=0

test_case() {
  description="$1"
  command="$2"
  expected="$3"
  output=$(eval "$command")
  if [[ "$output" == "$expected" ]]; then
    echo -e "[PASS] $description\n  Command: $command\n  Expected: $expected\n  Got:      $output\n"
    pass_count=$((pass_count+1))
  else
    echo -e "[FAIL] $description\n  Command: $command\n  Expected: $expected\n  Got:      $output\n"
    fail_count=$((fail_count+1))
  fi
}

echo "== Emma's POA (Time + Limit) checks =="
test_case "Emma can initiate payment under limit within timeframe (amount=3000, date=2025-03-15)" \
  "zed permission check poa:poa1 can_initiate_payment user:emma --caveat-context '{\"amount\":3000,\"now\":\"2025-03-15T00:00:00Z\"}'" true
test_case "Emma can initiate payment at limit within timeframe (amount=5000, date=2025-06-15)" \
  "zed permission check poa:poa1 can_initiate_payment user:emma --caveat-context '{\"amount\":5000,\"now\":\"2025-06-15T00:00:00Z\"}'" true
test_case "Emma cannot initiate payment above limit (amount=5001, date=2025-03-15)" \
  "zed permission check poa:poa1 can_initiate_payment user:emma --caveat-context '{\"amount\":5001,\"now\":\"2025-03-15T00:00:00Z\"}'" false
test_case "Emma cannot initiate payment before start date (amount=3000, date=2024-12-31)" \
  "zed permission check poa:poa1 can_initiate_payment user:emma --caveat-context '{\"amount\":3000,\"now\":\"2024-12-31T23:59:59Z\"}'" false
test_case "Emma cannot initiate payment after expiry (amount=3000, date=2025-07-01)" \
  "zed permission check poa:poa1 can_initiate_payment user:emma --caveat-context '{\"amount\":3000,\"now\":\"2025-07-01T00:00:00Z\"}'" false

echo "== Adi's Accountant Access checks =="
test_case "Adi can view transactions within timeframe (date=2025-02-15)" \
  "zed permission check account:acc456 can_view_transactions user:adi --caveat-context '{\"now\":\"2025-02-15T00:00:00Z\"}'" true
test_case "Adi can download statement within timeframe (date=2025-03-30)" \
  "zed permission check account:acc456 can_download_statement user:adi --caveat-context '{\"now\":\"2025-03-30T00:00:00Z\"}'" true
test_case "Adi cannot view transactions before start date (date=2024-12-31)" \
  "zed permission check account:acc456 can_view_transactions user:adi --caveat-context '{\"now\":\"2024-12-31T23:59:59Z\"}'" false
test_case "Adi cannot view transactions after expiry (date=2025-04-01)" \
  "zed permission check account:acc456 can_view_transactions user:adi --caveat-context '{\"now\":\"2025-04-01T00:00:00Z\"}'" false
test_case "Adi cannot download statement after expiry (date=2025-04-01)" \
  "zed permission check account:acc456 can_download_statement user:adi --caveat-context '{\"now\":\"2025-04-01T00:00:00Z\"}'" false
test_case "Adi cannot initiate payment (accountant restriction)" \
  "zed permission check account:acc456 can_initiate_payment user:adi --caveat-context '{\"now\":\"2025-02-15T00:00:00Z\"}'" false

echo "== POA Delegated Access checks =="
test_case "Emma can view transactions via POA on acc123 (within timeframe)" \
  "zed permission check account:acc123 can_view_transactions user:emma --caveat-context '{\"amount\":3000,\"now\":\"2025-03-15T00:00:00Z\"}'" true
test_case "Emma can download statement via POA on acc123 (within timeframe)" \
  "zed permission check account:acc123 can_download_statement user:emma --caveat-context '{\"amount\":3000,\"now\":\"2025-03-15T00:00:00Z\"}'" true
test_case "Emma cannot view transactions via POA on acc123 (after expiry)" \
  "zed permission check account:acc123 can_view_transactions user:emma --caveat-context '{\"amount\":3000,\"now\":\"2025-07-01T00:00:00Z\"}'" false

echo "== Direct and role-based checks =="
test_case "David can download statement for acc123 (owner)" \
  "zed permission check account:acc123 can_download_statement user:david" true
test_case "Emma can download statement for acc456 (owner)" \
  "zed permission check account:acc456 can_download_statement user:emma" true
test_case "Alice can_access org abc via finance_ops role (role membership)" \
  "zed permission check org:abc can_access user:alice" true
test_case "Bob cannot download statement for acc123 (no relationship)" \
  "zed permission check account:acc123 can_download_statement user:bob" false

echo "== Summary =="
echo "Passed: $pass_count"
echo "Failed: $fail_count"
if [[ $fail_count -eq 0 ]]; then
  echo "✅ All permission checks passed!"
else
  echo "❌ Some permission checks failed. Review output above."
fi 