# Go-Based Unit Tests for SpiceDB Model

This directory now includes a comprehensive Go-based unit test framework that replaces the bash-based integration tests while maintaining full compatibility with the SpiceDB entitlement system.

## Quick Start

```bash
# Complete setup and test
make setup test

# Or step by step:
make docker-up      # Start SpiceDB + PostgreSQL
make deps           # Install Go dependencies  
make migrate        # Run database migrations
make load-data      # Load schema and test data
make test           # Run Go unit tests
```

## Test Framework Architecture

### Core Components

1. **`test_client.go`** - SpiceDB client wrapper with test utilities
2. **`test_data.go`** - Test data management and setup functions
3. **`permissions_test.go`** - Comprehensive test suite using testify/suite
4. **`Makefile`** - Automation for setup, testing, and maintenance

### Test Coverage

The Go test suite covers all scenarios from the original bash tests:

#### Emma's POA (Power of Attorney) Tests ✅
- Time-bound access with amount limits (5 tests)
- Valid timeframe: Jan 1 - June 30, 2025
- Amount limit: $5,000
- Tests positive and negative scenarios for both time and amount constraints

#### Adi's Accountant Access Tests ✅ 
- Temporary accountant permissions (6 tests)
- Valid timeframe: Jan 1 - Mar 31, 2025
- Can view transactions and download statements
- Cannot initiate payments (role restriction)

#### Direct Ownership Tests ✅
- Account owners can access their accounts (2 tests)
- Non-owners cannot access accounts (1 test)

#### Role-Based Access Tests ✅
- Organization access through role membership (1 test)
- Alice → finance_ops → org:abc access chain

#### POA Delegation Tests ✅
- Emma's delegated access to David's account via POA (3 tests)
- Time-bound permissions through delegation chain

## Test Features

### Production-Like Setup
- **PostgreSQL 15** persistent storage (not in-memory)
- **Database migrations** run automatically
- **Health checks** ensure services are ready
- **Fresh data loading** before each test

### Comprehensive Test Isolation
- Each test method resets and reloads all data
- No test pollution or ordering dependencies
- Parallel-safe test execution

### Rich Assertions
- Uses `testify/suite` for structured testing
- Detailed error messages with context
- Clear pass/fail reporting

## File Structure

```
spicedb-model/
├── go.mod                    # Go module definition
├── test_client.go           # SpiceDB client wrapper
├── test_data.go             # Test data management
├── permissions_test.go      # Main test suite
├── Makefile                 # Automation commands
├── model.zaml               # SpiceDB schema
├── tuples.csv              # Test relationship data
└── spicedb-config/
    └── docker-compose.yml   # SpiceDB + PostgreSQL setup
```

## Running Tests

### Quick Commands

```bash
# Run all tests
make test

# Run with verbose output  
make test-verbose

# Run with race detection
make test-race

# Compare with bash tests
make test-all

# Reset data and test
make test-clean
```

### Individual Test Categories

```bash
# Run specific test patterns
go test -v -run="Emma.*POA" ./...
go test -v -run="Adi.*Accountant" ./...
go test -v -run="RoleBasedAccess" ./...
```

### Development Workflow

```bash
# Start development environment
make dev

# Make changes to tests...

# Run tests
make test

# Check data state
make inspect-all

# Reset if needed  
make reset-data
```

## Test Data Scenarios

### Financial Entitlement System

The tests simulate a realistic financial sector authorization model:

#### Accounts & Ownership
- **acc123** owned by David
- **acc456** owned by Emma  
- **acc789-792** owned by organization ABC

#### Power of Attorney (POA)
- Emma has POA on David's account (acc123)
- Time-bound: Valid until June 30, 2025
- Amount-limited: Max $5,000 per transaction

#### Temporary Access
- Adi has accountant access to Emma's account (acc456)
- Time-bound: Valid until March 31, 2025
- Read-only: Can view/download, cannot initiate payments

#### Organizational Access
- Alice is member of finance_ops role
- finance_ops role has access to organization ABC
- Transitive access: Alice → finance_ops → org ABC

## Advanced Usage

### Custom Test Data

Modify `test_data.go` to add new relationships:

```go
relationships := append(defaultRelationships, TestRelationship{
    Resource: "account:acc999",
    Relation: "owner", 
    Subject:  "user:bob",
    Caveat:   nil,
})
```

### Custom Test Cases

Add new test methods to `permissions_test.go`:

```go
func (suite *PermissionsTestSuite) TestCustomScenario() {
    result, err := suite.client.CheckPermission(
        "account:acc123",
        "can_initiate_payment", 
        "user:alice",
        nil,
    )
    
    require.NoError(suite.T(), err)
    assert.False(suite.T(), result)
}
```

### Database Inspection

```bash
# View current schema
make inspect-schema

# View current relationships  
make inspect-data

# View both
make inspect-all
```

## Comparison: Go vs Bash Tests

| Feature | Bash Tests | Go Tests |
|---------|------------|----------|
| **Language** | Shell scripting | Go |
| **Assertions** | String comparison | Rich testify assertions |
| **Isolation** | Manual reset script | Automatic per-test reset |
| **Parallelization** | None | Test-suite level |
| **IDE Support** | Limited | Full debugging/breakpoints |
| **CI/CD Integration** | Basic | Native Go test ecosystem |
| **Error Reporting** | Basic pass/fail | Detailed context & stack traces |
| **Performance** | ~15 test calls | Structured test execution |

## Performance

- **Total test time**: ~5-6 seconds for full suite
- **Individual tests**: ~50-100ms each  
- **Setup time**: ~2-3 seconds for data loading
- **Parallel capability**: Suite-level (can be enhanced)

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure Docker is running and containers are healthy
   ```bash
   make docker-up
   docker ps | grep spicedb
   ```

2. **Schema errors**: Check if migrations ran successfully
   ```bash
   make migrate
   make inspect-schema
   ```

3. **Test failures**: Verify data is loaded correctly
   ```bash
   make inspect-data
   make reset-data test
   ```

### Debug Mode

Enable verbose logging in tests:

```go
// In test_client.go, add logging
fmt.Printf("Checking permission: %s %s %s\n", resource, permission, subject)
```

## Future Enhancements

- [ ] Benchmark tests for performance analysis
- [ ] Property-based testing with random scenarios
- [ ] Integration with GitHub Actions CI/CD
- [ ] Test coverage reporting
- [ ] Multi-tenant test scenarios
- [ ] Load testing capabilities

---

This Go-based test framework provides a robust, maintainable, and scalable foundation for testing the SpiceDB entitlement system while maintaining full compatibility with the existing data model and use cases.