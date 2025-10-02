package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	pb "github.com/adityakumar/labs/go-entitlement-service/internal/pb/entitlement-service/proto"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "http://localhost:4000",
	}
}

func (c *Client) CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	start := time.Now()

	// Map the gRPC request to GraphQL query
	query, variables := c.buildGraphQLQuery(req)

	// Create GraphQL request
	gqlReq := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	// Serialize request
	reqBody, err := json.Marshal(gqlReq)
	if err != nil {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("failed to marshal GraphQL request: %v", err),
		}, nil
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("failed to create HTTP request: %v", err),
		}, nil
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-user-id", req.Actor)

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("failed to make GraphQL request: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("failed to read response body: %v", err),
		}, nil
	}

	// Parse GraphQL response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("failed to unmarshal GraphQL response: %v", err),
		}, nil
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		return &pb.PermissionResponse{
			HasPermission: false,
			ErrorMessage:  fmt.Sprintf("GraphQL errors: %v", gqlResp.Errors),
		}, nil
	}

	// Extract result from response
	hasPermission := c.extractPermissionResult(gqlResp.Data, req.Permission)

	responseTime := time.Since(start).Milliseconds()

	return &pb.PermissionResponse{
		HasPermission:  hasPermission,
		Permissionship: c.mapPermissionship(hasPermission),
		Implementation: pb.Implementation_IMPLEMENTATION_GRAPHQL,
		ResponseTimeMs: float64(responseTime),
		ErrorMessage:   "",
	}, nil
}

func (c *Client) buildGraphQLQuery(req *pb.PermissionRequest) (string, map[string]interface{}) {
	variables := map[string]interface{}{
		"accountId": req.Resource,
	}

	switch req.Permission {
	case "can_view_transactions":
		query := `
		query($accountId: ID!) {
			canViewTransactions(accountId: $accountId)
		}`
		return query, variables

	case "can_download_statement":
		query := `
		query($accountId: ID!) {
			canDownloadStatement(accountId: $accountId)
		}`
		return query, variables

	case "can_initiate_payment":
		// Extract amount from context if available
		amount := 1000.0 // default
		if amtStr, ok := req.Context["amount"]; ok {
			// Try to parse as float64 directly
			if amt, err := strconv.ParseFloat(amtStr, 64); err == nil {
				amount = amt
			}
		}
		variables["amount"] = amount

		query := `
		query($accountId: ID!, $amount: Float!) {
			canInitiatePayment(accountId: $accountId, amount: $amount)
		}`
		return query, variables

	case "can_access":
		query := `
		query($accountId: ID!) {
			canAccess(accountId: $accountId)
		}`
		return query, variables

	default:
		// Default to access check for unknown permissions
		query := `
		query($accountId: ID!) {
			canAccess(accountId: $accountId)
		}`
		return query, variables
	}
}

func (c *Client) extractPermissionResult(data map[string]interface{}, permission string) bool {
	// Extract the first boolean value from the response
	for _, value := range data {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return false
}

func (c *Client) mapPermissionship(hasPermission bool) int32 {
	if hasPermission {
		return 2 // PERMISSIONSHIP_HAS_PERMISSION
	}
	return 1 // PERMISSIONSHIP_NO_PERMISSION
}
