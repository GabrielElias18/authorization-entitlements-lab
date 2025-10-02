package service

import (
	"context"
	"fmt"

	pb "github.com/adityakumar/labs/go-entitlement-service/internal/pb/entitlement-service/proto"
)

type SpiceDBClient interface {
	CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error)
}

type Neo4jClient interface {
	CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error)
}

type GraphQLClient interface {
	CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error)
}

type Service struct {
	Spice   SpiceDBClient
	Neo4j   Neo4jClient
	GraphQL GraphQLClient
}

func NewService(spice SpiceDBClient, neo4j Neo4jClient, gql GraphQLClient) *Service {
	return &Service{Spice: spice, Neo4j: neo4j, GraphQL: gql}
}

// CheckPermission routes the request to the appropriate backend based on implementation
func (s *Service) CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	// Default to SpiceDB if no specific implementation is requested
	implementation := pb.Implementation_IMPLEMENTATION_SPICEDB

	// Check if implementation is specified in context
	if implStr, ok := req.Context["implementation"]; ok {
		switch implStr {
		case "spicedb":
			implementation = pb.Implementation_IMPLEMENTATION_SPICEDB
		case "neo4j":
			implementation = pb.Implementation_IMPLEMENTATION_NEO4J
		case "graphql":
			implementation = pb.Implementation_IMPLEMENTATION_GRAPHQL
		case "both":
			implementation = pb.Implementation_IMPLEMENTATION_BOTH
		}
	}

	switch implementation {
	case pb.Implementation_IMPLEMENTATION_NEO4J:
		if s.Neo4j == nil {
			return &pb.PermissionResponse{
				HasPermission:  false,
				Implementation: pb.Implementation_IMPLEMENTATION_NEO4J,
				ErrorMessage:   "Neo4j backend not available - please ensure Neo4j is running on localhost:7687",
			}, nil
		}
		return s.Neo4j.CheckPermission(ctx, req)
	case pb.Implementation_IMPLEMENTATION_GRAPHQL:
		if s.GraphQL == nil {
			return &pb.PermissionResponse{
				HasPermission:  false,
				Implementation: pb.Implementation_IMPLEMENTATION_GRAPHQL,
				ErrorMessage:   "GraphQL backend not available - please ensure GraphQL is running on localhost:4000",
			}, nil
		}
		return s.GraphQL.CheckPermission(ctx, req)
	case pb.Implementation_IMPLEMENTATION_BOTH:
		return s.checkBothImplementations(ctx, req)
	case pb.Implementation_IMPLEMENTATION_SPICEDB:
		fallthrough
	default:
		if s.Spice == nil {
			return &pb.PermissionResponse{
				HasPermission:  false,
				Implementation: pb.Implementation_IMPLEMENTATION_SPICEDB,
				ErrorMessage:   "SpiceDB backend not available - please ensure SpiceDB is running on localhost:50051",
			}, nil
		}
		return s.Spice.CheckPermission(ctx, req)
	}
}

// checkBothImplementations runs the same request against both SpiceDB and Neo4j for comparison
func (s *Service) checkBothImplementations(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	var spiceResult, neo4jResult *pb.PermissionResponse
	var spiceErr, neo4jErr error

	// Run both implementations concurrently
	spiceChan := make(chan struct{})
	neo4jChan := make(chan struct{})

	go func() {
		defer close(spiceChan)
		if s.Spice != nil {
			spiceResult, spiceErr = s.Spice.CheckPermission(ctx, req)
		} else {
			spiceErr = fmt.Errorf("SpiceDB backend not available")
		}
	}()

	go func() {
		defer close(neo4jChan)
		if s.Neo4j != nil {
			neo4jResult, neo4jErr = s.Neo4j.CheckPermission(ctx, req)
		} else {
			neo4jErr = fmt.Errorf("Neo4j backend not available")
		}
	}()

	// Wait for both to complete
	<-spiceChan
	<-neo4jChan

	// If both have errors, return comparison error
	if spiceErr != nil && neo4jErr != nil {
		return &pb.PermissionResponse{
			HasPermission:  false,
			Implementation: pb.Implementation_IMPLEMENTATION_BOTH,
			ErrorMessage:   fmt.Sprintf("Both backends unavailable - SpiceDB: %v, Neo4j: %v", spiceErr, neo4jErr),
		}, nil
	}

	// If one has an error, log it but continue with the working one
	if spiceErr != nil {
		fmt.Printf("SpiceDB error (using Neo4j only): %v\n", spiceErr)
		if neo4jResult != nil {
			neo4jResult.Implementation = pb.Implementation_IMPLEMENTATION_BOTH
			neo4jResult.ErrorMessage = fmt.Sprintf("SpiceDB unavailable, using Neo4j only: %s", neo4jResult.ErrorMessage)
			return neo4jResult, nil
		}
	}

	if neo4jErr != nil {
		fmt.Printf("Neo4j error (using SpiceDB only): %v\n", neo4jErr)
		if spiceResult != nil {
			spiceResult.Implementation = pb.Implementation_IMPLEMENTATION_BOTH
			spiceResult.ErrorMessage = fmt.Sprintf("Neo4j unavailable, using SpiceDB only: %s", spiceResult.ErrorMessage)
			return spiceResult, nil
		}
	}

	// Both succeeded - compare results and return SpiceDB result with comparison info
	if spiceResult != nil && neo4jResult != nil {
		// Log comparison for debugging
		if spiceResult.HasPermission != neo4jResult.HasPermission {
			fmt.Printf("MISMATCH: SpiceDB=%t, Neo4j=%t for actor=%s, resource=%s, permission=%s\n",
				spiceResult.HasPermission, neo4jResult.HasPermission,
				req.Actor, req.Resource, req.Permission)
		} else {
			fmt.Printf("MATCH: Both backends returned %t for actor=%s, resource=%s, permission=%s\n",
				spiceResult.HasPermission, req.Actor, req.Resource, req.Permission)
		}

		// Return SpiceDB result but mark as BOTH implementation
		spiceResult.Implementation = pb.Implementation_IMPLEMENTATION_BOTH
		return spiceResult, nil
	}

	return nil, fmt.Errorf("unexpected state in both implementations check")
}
