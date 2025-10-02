package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PermissionsTestSuite contains all permission tests
type PermissionsTestSuite struct {
	suite.Suite
	client *TestClient
}

// SetupSuite runs once before all tests
func (suite *PermissionsTestSuite) SetupSuite() {
	client, err := NewTestClient()
	require.NoError(suite.T(), err, "Failed to create test client")
	suite.client = client

	// Wait a moment for SpiceDB to be ready
	time.Sleep(2 * time.Second)
}

// TearDownSuite runs once after all tests
func (suite *PermissionsTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.client.Close()
	}
}

// SetupTest runs before each test
func (suite *PermissionsTestSuite) SetupTest() {
	err := SetupTestData(suite.client)
	require.NoError(suite.T(), err, "Failed to setup test data")
}

// TestEmmasPOATimeAndLimitChecks tests Emma's Power of Attorney permissions with time and limit constraints
func (suite *PermissionsTestSuite) TestEmmasPOATimeAndLimitChecks() {
	testCases := []struct {
		name     string
		amount   int
		now      string
		expected bool
	}{
		{
			name:     "Emma can initiate payment under limit within timeframe (amount=3000, date=2025-03-15)",
			amount:   3000,
			now:      "2025-03-15T00:00:00Z",
			expected: true,
		},
		{
			name:     "Emma can initiate payment at limit within timeframe (amount=5000, date=2025-06-15)",
			amount:   5000,
			now:      "2025-06-15T00:00:00Z",
			expected: true,
		},
		{
			name:     "Emma cannot initiate payment above limit (amount=5001, date=2025-03-15)",
			amount:   5001,
			now:      "2025-03-15T00:00:00Z",
			expected: false,
		},
		{
			name:     "Emma cannot initiate payment before start date (amount=3000, date=2024-12-31)",
			amount:   3000,
			now:      "2024-12-31T23:59:59Z",
			expected: false,
		},
		{
			name:     "Emma cannot initiate payment after expiry (amount=3000, date=2025-07-01)",
			amount:   3000,
			now:      "2025-07-01T00:00:00Z",
			expected: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			caveatContext := map[string]interface{}{
				"amount": tc.amount,
				"now":    tc.now,
			}

			result, err := suite.client.CheckPermission(
				"poa:poa1",
				"can_initiate_payment",
				"user:emma",
				caveatContext,
			)

			require.NoError(suite.T(), err, "Permission check should not error")
			assert.Equal(suite.T(), tc.expected, result, "Permission result should match expected")
		})
	}
}

// TestAdisAccountantAccessChecks tests Adi's temporary accountant access permissions
func (suite *PermissionsTestSuite) TestAdisAccountantAccessChecks() {
	testCases := []struct {
		name       string
		permission string
		now        string
		expected   bool
	}{
		{
			name:       "Adi can view transactions within timeframe (date=2025-02-15)",
			permission: "can_view_transactions",
			now:        "2025-02-15T00:00:00Z",
			expected:   true,
		},
		{
			name:       "Adi can download statement within timeframe (date=2025-03-30)",
			permission: "can_download_statement",
			now:        "2025-03-30T00:00:00Z",
			expected:   true,
		},
		{
			name:       "Adi cannot view transactions before start date (date=2024-12-31)",
			permission: "can_view_transactions",
			now:        "2024-12-31T23:59:59Z",
			expected:   false,
		},
		{
			name:       "Adi cannot view transactions after expiry (date=2025-04-01)",
			permission: "can_view_transactions",
			now:        "2025-04-01T00:00:00Z",
			expected:   false,
		},
		{
			name:       "Adi cannot download statement after expiry (date=2025-04-01)",
			permission: "can_download_statement",
			now:        "2025-04-01T00:00:00Z",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			caveatContext := map[string]interface{}{
				"now": tc.now,
			}

			result, err := suite.client.CheckPermission(
				"account:acc456",
				tc.permission,
				"user:adi",
				caveatContext,
			)

			require.NoError(suite.T(), err, "Permission check should not error")
			assert.Equal(suite.T(), tc.expected, result, "Permission result should match expected")
		})
	}
}

// TestAdisPaymentRestriction tests that accountants cannot initiate payments
func (suite *PermissionsTestSuite) TestAdisPaymentRestriction() {
	suite.Run("Adi cannot initiate payment (accountant restriction)", func() {
		caveatContext := map[string]interface{}{
			"now": "2025-02-15T00:00:00Z",
		}

		result, err := suite.client.CheckPermission(
			"account:acc456",
			"can_initiate_payment",
			"user:adi",
			caveatContext,
		)

		require.NoError(suite.T(), err, "Permission check should not error")
		assert.False(suite.T(), result, "Adi should not be able to initiate payments")
	})
}

// TestDirectOwnershipPermissions tests direct account ownership permissions
func (suite *PermissionsTestSuite) TestDirectOwnershipPermissions() {
	testCases := []struct {
		name       string
		account    string
		permission string
		user       string
		expected   bool
	}{
		{
			name:       "David can download statement for acc123 (owner)",
			account:    "account:acc123",
			permission: "can_download_statement",
			user:       "user:david",
			expected:   true,
		},
		{
			name:       "Emma can download statement for acc456 (owner)",
			account:    "account:acc456",
			permission: "can_download_statement",
			user:       "user:emma",
			expected:   true,
		},
		{
			name:       "Bob cannot download statement for acc123 (no relationship)",
			account:    "account:acc123",
			permission: "can_download_statement",
			user:       "user:bob",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.client.CheckPermission(
				tc.account,
				tc.permission,
				tc.user,
				nil, // No caveat context needed for direct ownership
			)

			require.NoError(suite.T(), err, "Permission check should not error")
			assert.Equal(suite.T(), tc.expected, result, "Permission result should match expected")
		})
	}
}

// TestRoleBasedAccess tests role-based organization access
func (suite *PermissionsTestSuite) TestRoleBasedAccess() {
	suite.Run("Alice can access org abc via finance_ops role", func() {
		result, err := suite.client.CheckPermission(
			"org:abc",
			"can_access",
			"user:alice",
			nil, // No caveat context needed for role-based access
		)

		require.NoError(suite.T(), err, "Permission check should not error")
		assert.True(suite.T(), result, "Alice should have access to org abc through role membership")
	})
}

// TestPOAViewAndDownloadPermissions tests POA permissions for viewing and downloading via delegated access
func (suite *PermissionsTestSuite) TestPOAViewAndDownloadPermissions() {
	testCases := []struct {
		name       string
		permission string
		amount     int
		now        string
		expected   bool
	}{
		{
			name:       "Emma can view transactions via POA on acc123 (within timeframe)",
			permission: "can_view_transactions",
			amount:     3000,
			now:        "2025-03-15T00:00:00Z",
			expected:   true,
		},
		{
			name:       "Emma can download statement via POA on acc123 (within timeframe)",
			permission: "can_download_statement", 
			amount:     3000,
			now:        "2025-03-15T00:00:00Z",
			expected:   true,
		},
		{
			name:       "Emma cannot view transactions via POA on acc123 (after expiry)",
			permission: "can_view_transactions",
			amount:     3000,
			now:        "2025-07-01T00:00:00Z",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			caveatContext := map[string]interface{}{
				"amount": tc.amount,
				"now":    tc.now,
			}

			// Emma has delegated access to acc123 through poa:poa1
			result, err := suite.client.CheckPermission(
				"account:acc123", // David's account where Emma has POA
				tc.permission,
				"user:emma",
				caveatContext,
			)

			require.NoError(suite.T(), err, "Permission check should not error")
			assert.Equal(suite.T(), tc.expected, result, "Permission result should match expected")
		})
	}
}

// TestRunAllPermissionTests runs the test suite
func TestRunAllPermissionTests(t *testing.T) {
	suite.Run(t, new(PermissionsTestSuite))
}