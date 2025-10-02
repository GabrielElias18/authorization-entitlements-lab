# SpiceDB Model:  Entitlements Engine

This module implements an authorization system using SpiceDB, demonstrating common access control patterns for resource management with delegation and contextual rules.

This directory contains the schema, data, and automation for loading and testing entitlement/authorization models in SpiceDB.

## Sample Use Cases Covered

Use entitlement checks to determine whether an end user is authorized to perform the following functions:

- View Transactions 
- Initiate Transaction
- Download account statements
- Make Transaction within Limits
- Have Delegated access (from another user)
- HaveTime Bound Access
- Access and delegation by organization membership

## Core SpiceDB Concepts

### 🔗 **Relationships** - Who connects to what
**Think of it as:** *Bank account ownership or family relationships*

Relationships are stored as tuples: `resource:id relation subject:id`
```
account:acc123 owner user:david
account:acc123 delegated_access poa:poa1  
poa:poa1 delegate_with_time_and_limit user:emma
```

### 📋 **Permissions** - What someone can do  
**Think of it as:** *Having the right key opens the right door*

Permissions are computed from relationships using schema rules. David can download his statements because he owns the account, while Bob (a stranger) cannot.
```yaml
permission can_download_statement = owner + delegated_access->can_download_statement
```

### ⚡ **Caveats** - "Yes, but only if..." conditions
**Think of it as:** *Credit card with spending limits and expiration dates*  

Caveats add dynamic conditions evaluated at check time. Emma's delegated authority for David is valid until June 30, 2025, with a $5000 transaction limit.
```yaml
caveat within_time_and_limit(max_amount int, amount int, start timestamp, end timestamp, now timestamp) {
  amount <= max_amount && now >= start && now <= end
}
```

### 🔄 **Delegation** - Authority passed through chains
**Think of it as:** *Power of Attorney - David grants Emma authority over his account*

Complex authorization flows through multiple objects. Emma can access David's account because David granted her right to act on his behalf, and along with specific permissions with time/amount restrictions.
```
David's Account → POA Object → Emma + Caveats = Conditional Access
```

## Quick Start

### Using Makefile (Recommended)
```bash
# Complete setup and run all tests
make setup test-all

# Or step by step:
make docker-up      # Start SpiceDB + PostgreSQL
make migrate        # Run database migrations  
make load-data      # Load schema and test data
make test-all       # Run both Go and bash tests
```

### Manual Setup (Original)
```bash
# 1. Setup SpiceDB with schema and data
bash setup.sh

# 2. Run permission tests
bash run_permission_checks.sh

# 3. Reset data (if needed)
bash reset.sh
```

## Contents
- `model.zaml` — SpiceDB schema (Zanzibar/SpiceDB language)
- `tuples.csv` — Relationship and caveat data for bulk import
- `setup.sh` — One-command setup: starts SpiceDB, loads schema and data
- `reset.sh` — Clear and reload data (keeps SpiceDB running)
- `run_permission_checks.sh` — Bash-based integration tests (19 test cases)
- `spicedb-config/` — Docker Compose configuration for SpiceDB + PostgreSQL
- `tests/go/` — Go-based unit test framework with comprehensive test suite
- `Makefile` — Automation for setup, testing, and development

## Workflow

### 1. Initial Setup
```bash
bash setup.sh
```
This single command will:
- Start SpiceDB + PostgreSQL containers
- Install zed CLI (if needed)
- Run database migrations
- Load schema from `model.zaml`
- Load relationships from `tuples.csv`

### 2. Run Tests
```bash
bash run_permission_checks.sh
```
This will check all key permissions and caveat scenarios, printing pass/fail for each test.

### 3. Reset Data (Optional)
```bash
bash reset.sh
```
Clear all data and reload from files (useful after schema changes).

### 4. Caveat Support
- Caveated relationships are supported via the `caveat` column in `tuples.csv`
- Only static parameters (e.g., `max_amount`, `start`, `end`) should be stored in the relationship context
- Dynamic parameters (e.g., `amount`, `now`) are provided at check time in the test script

## Requirements
- [Docker](https://docker.com/) (for SpiceDB + PostgreSQL containers)
- [Go 1.21+](https://golang.org/) (for Go-based unit tests)
- [Homebrew](https://brew.sh/) (for zed CLI installation)
- Bash shell

The setup script will automatically install the `zed` CLI if it's not present.

## Testing

This project includes two comprehensive test suites:

### Bash Integration Tests 
- **18 test cases** covering all permission scenarios
- Direct SpiceDB CLI integration
- Simple pass/fail reporting
- Run with: `make test-bash` or `bash run_permission_checks.sh`

### Go Unit Tests 
- **24 individual test cases** with rich assertions
- Structured testing using testify/suite
- Automatic data setup/teardown per test
- Production-like workflow with persistent PostgreSQL
- Run with: `make test` or `cd tests/go && go test -v`

### Combined Testing
```bash
make test-all    # Run both test suites
make test-clean  # Reset data and run Go tests
```

## Extending

### For Bash Tests
- Add new relationships or caveats to `tuples.csv`
- Add new test cases to `run_permission_checks.sh`
- Update the schema in `model.zaml` as needed

### For Go Tests
- Modify `tests/go/test_data.go` to add new relationships
- Add new test methods in `tests/go/permissions_test.go`
- See `tests/go/README_GO_TESTS.md` for detailed Go test documentation

## Example: Adding a New Caveated Relationship
Add to `tuples.csv`:
```
poa:poa3,delegate_with_limit,user:bob,under_limit:{"max_amount":1000}
```
Then reload with:
```bash
bash reset.sh
```

## Example: Adding a New Test
Add to `run_permission_checks.sh`:
```bash
test_case "Bob can initiate transaction under new limit (amount=500)" \
  "zed permission check poa:poa3 can_initiate_payment user:bob --caveat-context '{\"amount\":500}'" true
```

---
For more details, see the main project README or the [SpiceDB documentation](https://authzed.com/docs/spicedb/). 