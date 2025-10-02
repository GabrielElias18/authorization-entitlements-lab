package main

import (
	"io/ioutil"
	"path/filepath"
)

// TestData holds all test relationships and schema
type TestData struct {
	Schema        string
	Relationships []TestRelationship
}

// TestRelationship represents a relationship to be loaded
type TestRelationship struct {
	Resource string
	Relation string
	Subject  string
	Caveat   map[string]interface{}
}

// LoadTestData loads schema and relationships for testing
func LoadTestData() (*TestData, error) {
	// Load schema from model.zaml (go up two directories to reach project root)
	schemaPath := filepath.Join("..", "..", "model.zaml")
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	// Define test relationships matching tuples.csv
	relationships := []TestRelationship{
		{
			Resource: "account:acc123",
			Relation: "owner",
			Subject:  "user:david",
			Caveat:   nil,
		},
		{
			Resource: "account:acc456",
			Relation: "owner",
			Subject:  "user:emma",
			Caveat:   nil,
		},
		{
			Resource: "account:acc123",
			Relation: "delegated_access",
			Subject:  "poa:poa1",
			Caveat:   nil,
		},
		{
			Resource: "poa:poa1",
			Relation: "delegate_with_time_and_limit",
			Subject:  "user:emma",
			Caveat: map[string]interface{}{
				"max_amount": 5000,
				"start":      "2025-01-01T00:00:00Z",
				"end":        "2025-06-30T23:59:59Z",
			},
		},
		{
			Resource: "account:acc456",
			Relation: "accountant_access",
			Subject:  "user:adi",
			Caveat: map[string]interface{}{
				"start": "2025-01-01T00:00:00Z",
				"end":   "2025-03-31T23:59:59Z",
			},
		},
		{
			Resource: "account:acc789",
			Relation: "owner",
			Subject:  "org:abc",
			Caveat:   nil,
		},
		{
			Resource: "account:acc790",
			Relation: "owner",
			Subject:  "org:abc",
			Caveat:   nil,
		},
		{
			Resource: "account:acc791",
			Relation: "owner",
			Subject:  "org:abc",
			Caveat:   nil,
		},
		{
			Resource: "account:acc792",
			Relation: "owner",
			Subject:  "org:abc",
			Caveat:   nil,
		},
		{
			Resource: "role:finance_ops",
			Relation: "member",
			Subject:  "user:alice",
			Caveat:   nil,
		},
		{
			Resource: "org:abc",
			Relation: "member",
			Subject:  "role:finance_ops",
			Caveat:   nil,
		},
	}

	return &TestData{
		Schema:        string(schemaBytes),
		Relationships: relationships,
	}, nil
}

// SetupTestData clears existing data and loads fresh test data
func SetupTestData(client *TestClient) error {
	// Load test data
	testData, err := LoadTestData()
	if err != nil {
		return err
	}

	// Clear all existing data
	if err := client.ClearAllData(); err != nil {
		return err
	}

	// Load schema
	if err := client.LoadSchema(testData.Schema); err != nil {
		return err
	}

	// Load all relationships
	for _, rel := range testData.Relationships {
		if err := client.CreateRelationship(rel.Resource, rel.Relation, rel.Subject, rel.Caveat); err != nil {
			return err
		}
	}

	return nil
}