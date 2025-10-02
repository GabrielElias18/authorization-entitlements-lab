package main

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestClient wraps the SpiceDB client with test utilities
type TestClient struct {
	client *authzed.Client
	ctx    context.Context
}

// NewTestClient creates a new test client connected to SpiceDB
func NewTestClient() (*TestClient, error) {
	// Configure connection with preshared key
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer dev-key")
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	}

	client, err := authzed.NewClient(
		"localhost:50051",
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SpiceDB client: %w", err)
	}

	return &TestClient{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// Close closes the client connection
func (tc *TestClient) Close() error {
	// authzed.Client doesn't have a Close method in this version
	return nil
}

// ClearAllData clears all relationships from SpiceDB
func (tc *TestClient) ClearAllData() error {
	objectTypes := []string{"user", "org", "role", "account", "poa"}
	
	for _, objectType := range objectTypes {
		filter := &v1.RelationshipFilter{
			ResourceType: objectType,
		}
		
		// Delete relationships in batches
		req := &v1.DeleteRelationshipsRequest{
			RelationshipFilter: filter,
		}
		
		_, err := tc.client.DeleteRelationships(tc.ctx, req)
		if err != nil {
			return fmt.Errorf("failed to delete %s relationships: %w", objectType, err)
		}
		
		// Small delay between deletions
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}

// LoadSchema loads the schema from model.zaml content
func (tc *TestClient) LoadSchema(schemaContent string) error {
	req := &v1.WriteSchemaRequest{
		Schema: schemaContent,
	}
	
	_, err := tc.client.WriteSchema(tc.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}
	
	return nil
}

// CreateRelationship creates a single relationship
func (tc *TestClient) CreateRelationship(resource, relation, subject string, caveat map[string]interface{}) error {
	// Parse resource
	resourceParts := parseObjectRef(resource)
	if len(resourceParts) != 2 {
		return fmt.Errorf("invalid resource format: %s", resource)
	}
	
	// Parse subject
	subjectParts := parseObjectRef(subject)
	if len(subjectParts) != 2 {
		return fmt.Errorf("invalid subject format: %s", subject)
	}
	
	relationship := &v1.Relationship{
		Resource: &v1.ObjectReference{
			ObjectType: resourceParts[0],
			ObjectId:   resourceParts[1],
		},
		Relation: relation,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subjectParts[0],
				ObjectId:   subjectParts[1],
			},
		},
	}
	
	// Add caveat if provided
	if caveat != nil && len(caveat) > 0 {
		relationship.OptionalCaveat = &v1.ContextualizedCaveat{
			CaveatName: getCaveatName(relation),
			Context:    convertToStruct(caveat),
		}
	}
	
	req := &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: relationship,
			},
		},
	}
	
	_, err := tc.client.WriteRelationships(tc.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create relationship %s %s %s: %w", resource, relation, subject, err)
	}
	
	return nil
}

// CheckPermission checks if a subject has permission on a resource
func (tc *TestClient) CheckPermission(resource, permission, subject string, caveatContext map[string]interface{}) (bool, error) {
	// Parse resource and subject
	resourceParts := parseObjectRef(resource)
	if len(resourceParts) != 2 {
		return false, fmt.Errorf("invalid resource format: %s", resource)
	}
	
	subjectParts := parseObjectRef(subject)
	if len(subjectParts) != 2 {
		return false, fmt.Errorf("invalid subject format: %s", subject)
	}
	
	req := &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: resourceParts[0],
			ObjectId:   resourceParts[1],
		},
		Permission: permission,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subjectParts[0],
				ObjectId:   subjectParts[1],
			},
		},
	}
	
	// Add caveat context if provided
	if caveatContext != nil && len(caveatContext) > 0 {
		req.Context = convertToStruct(caveatContext)
	}
	
	resp, err := tc.client.CheckPermission(tc.ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	
	return resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

// Helper functions

func parseObjectRef(objectRef string) []string {
	// Split "type:id" format
	for i, r := range objectRef {
		if r == ':' {
			return []string{objectRef[:i], objectRef[i+1:]}
		}
	}
	return []string{objectRef}
}

func getCaveatName(relation string) string {
	// Map relations to their caveat names based on the schema
	caveatMap := map[string]string{
		"delegate_with_limit":          "under_limit",
		"delegate_with_time":           "within_active_range", 
		"delegate_with_time_and_limit": "within_time_and_limit",
		"accountant_access":            "within_active_range",
	}
	
	if caveat, exists := caveatMap[relation]; exists {
		return caveat
	}
	
	return ""
}

func convertToStruct(data map[string]interface{}) *structpb.Struct {
	s, err := structpb.NewStruct(data)
	if err != nil {
		// Fallback to empty struct if conversion fails
		return &structpb.Struct{}
	}
	return s
}