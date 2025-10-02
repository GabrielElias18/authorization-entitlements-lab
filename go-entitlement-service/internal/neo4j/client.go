package neo4j

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/adityakumar/labs/go-entitlement-service/internal/pb/entitlement-service/proto"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Client struct {
	driver neo4j.Driver
}

func NewClient(uri, username, password string) (*Client, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Test the connection
	if err := driver.VerifyConnectivity(); err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Client{driver: driver}, nil
}

func (c *Client) Close() error {
	return c.driver.Close()
}

func (c *Client) CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	startTime := time.Now()

	// Map permission names to Cypher queries
	query, err := c.getPermissionQuery(req.Permission)
	if err != nil {
		return nil, err
	}

	// Prepare parameters
	params := map[string]interface{}{
		"userId":    req.Actor,
		"accountId": req.Resource,
	}

	// Add test date from context if provided, otherwise use current time
	testDate := time.Now().Format("2006-01-02T15:04:05")
	if dateStr, ok := req.Context["test_date"]; ok {
		testDate = dateStr
	}
	params["now"] = testDate

	// Add amount parameter for payment queries
	if req.Permission == "can_initiate_payment" || req.Permission == "write" {
		amount := 1000.0 // default
		if amtStr, ok := req.Context["amount"]; ok {
			if amt, err := strconv.ParseFloat(amtStr, 64); err == nil {
				amount = amt
			}
		}
		params["amount"] = amount
	}

	// Execute query
	session := c.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.Run(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Neo4j query: %w", err)
	}

	// Get the first record
	record, err := result.Single()
	if err != nil {
		responseTime := time.Since(startTime).Milliseconds()
		return &pb.PermissionResponse{
			HasPermission:  false,
			Permissionship: int32(pb.Permissionship_PERMISSIONSHIP_NO_PERMISSION),
			Implementation: pb.Implementation_IMPLEMENTATION_NEO4J,
			ResponseTimeMs: float64(responseTime),
		}, nil
	}

	// Extract the permission result
	hasPermission, ok := record.Get("result")
	if !ok {
		return nil, fmt.Errorf("query result does not contain 'result' field")
	}

	// Convert to boolean
	var permission bool
	switch v := hasPermission.(type) {
	case bool:
		permission = v
	case int64:
		permission = v != 0
	case float64:
		permission = v != 0
	default:
		return nil, fmt.Errorf("unexpected result type: %T", hasPermission)
	}

	// Map to permissionship
	var permissionship pb.Permissionship
	if permission {
		permissionship = pb.Permissionship_PERMISSIONSHIP_HAS_PERMISSION
	} else {
		permissionship = pb.Permissionship_PERMISSIONSHIP_NO_PERMISSION
	}

	responseTime := time.Since(startTime).Milliseconds()

	return &pb.PermissionResponse{
		HasPermission:  permission,
		Permissionship: int32(permissionship),
		Implementation: pb.Implementation_IMPLEMENTATION_NEO4J,
		ResponseTimeMs: float64(responseTime),
	}, nil
}

func (c *Client) getPermissionQuery(permission string) (string, error) {
	switch permission {
	case "read", "can_view_transactions":
		return `
			OPTIONAL MATCH (user:User {id: $userId})
			OPTIONAL MATCH (account:Account {id: $accountId})
			WITH user, account
			OPTIONAL MATCH (user)-[owns:OWNS]->(account)
			WITH user, account, owns IS NOT NULL as direct_owner
			OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
			WHERE datetime($now) >= datetime(poa.starts_at) AND datetime($now) <= datetime(poa.expires_at)
			WITH user, account, direct_owner, count(poa) as valid_poa
			OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
			WITH user, account, direct_owner, valid_poa, count(role) as role_access
			OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
			WHERE datetime($now) >= datetime(acc.starts_at) AND datetime($now) <= datetime(acc.expires_at)
			WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
			RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as result
		`, nil

	case "write", "can_initiate_payment":
		return `
			OPTIONAL MATCH (user:User {id: $userId})
			OPTIONAL MATCH (account:Account {id: $accountId})
			WITH user, account
			OPTIONAL MATCH (user)-[owns:OWNS]->(account)
			WITH user, account, owns IS NOT NULL as direct_owner
			OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
			WHERE datetime($now) >= datetime(poa.starts_at) AND datetime($now) <= datetime(poa.expires_at) AND ($amount <= poa.payment_limit OR poa.payment_limit IS NULL)
			WITH user, account, direct_owner, count(poa) as valid_poa
			OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
			WITH user, account, direct_owner, valid_poa, count(role) as role_access
			RETURN (direct_owner OR valid_poa > 0 OR role_access > 0) as result
		`, nil

	case "can_download_statement":
		return `
			OPTIONAL MATCH (user:User {id: $userId})
			OPTIONAL MATCH (account:Account {id: $accountId})
			WITH user, account
			OPTIONAL MATCH (user)-[owns:OWNS]->(account)
			WITH user, account, owns IS NOT NULL as direct_owner
			OPTIONAL MATCH (user)-[poa:HAS_POA]->(account)
			WHERE datetime($now) >= datetime(poa.starts_at) AND datetime($now) <= datetime(poa.expires_at)
			WITH user, account, direct_owner, count(poa) as valid_poa
			OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
			WITH user, account, direct_owner, valid_poa, count(role) as role_access
			OPTIONAL MATCH (user)-[acc:HAS_ACCOUNTANT_ACCESS]->(account)
			WHERE datetime($now) >= datetime(acc.starts_at) AND datetime($now) <= datetime(acc.expires_at)
			WITH user, account, direct_owner, valid_poa, role_access, count(acc) as accountant_access
			RETURN (direct_owner OR valid_poa > 0 OR role_access > 0 OR accountant_access > 0) as result
		`, nil

	case "can_access":
		return `
			OPTIONAL MATCH (user:User {id: $userId})
			OPTIONAL MATCH (account:Account {id: $accountId})
			WITH user, account
			OPTIONAL MATCH (user)-[owns:OWNS]->(account)
			WITH user, account, owns IS NOT NULL as direct_owner
			OPTIONAL MATCH (user)-[:HAS_ROLE]->(role:Role)-[:MEMBER_OF]->(org:Org)-[:OWNS]->(account)
			WITH user, account, direct_owner, count(role) as role_access
			RETURN (direct_owner OR role_access > 0) as result
		`, nil

	default:
		return "", fmt.Errorf("unsupported permission: %s", permission)
	}
}
