package spicedb

import (
	"context"
	"fmt"

	pb "github.com/adityakumar/labs/go-entitlement-service/internal/pb/entitlement-service/proto"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	client *authzed.Client
}

func NewClient() *Client {
	// Add preshared key authentication
	authInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer dev-key")
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	cli, err := authzed.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create SpiceDB client: %v", err))
	}
	return &Client{client: cli}
}

func (c *Client) CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	// Map proto request to SpiceDB format
	checkReq := &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: "account",
			ObjectId:   req.GetResource(),
		},
		Permission: req.GetPermission(),
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   req.GetActor(),
			},
		},
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{
				FullyConsistent: true,
			},
		},
	}

	// Add context if provided
	if req.Context != nil && len(req.Context) > 0 {
		// Convert map[string]string to map[string]interface{} for structpb
		contextMap := make(map[string]interface{})
		for key, value := range req.Context {
			contextMap[key] = value
		}
		contextStruct, err := structpb.NewStruct(contextMap)
		if err != nil {
			return &pb.PermissionResponse{HasPermission: false, ErrorMessage: fmt.Sprintf("failed to create context: %v", err)}, err
		}
		checkReq.Context = contextStruct
	}

	// Log the request for debugging
	fmt.Printf("SpiceDB Request: %+v\n", checkReq)

	resp, err := c.client.CheckPermission(ctx, checkReq)
	if err != nil {
		return &pb.PermissionResponse{HasPermission: false, ErrorMessage: err.Error()}, err
	}

	// Log the response for debugging
	fmt.Printf("SpiceDB Response: %+v\n", resp)

	// Handle different permissionship values
	var hasPermission bool
	switch resp.Permissionship {
	case v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION:
		hasPermission = true
	case v1.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION:
		// Conditional permission means the caveat conditions were met with the provided context
		// This is equivalent to having permission
		hasPermission = true
	case v1.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION:
		hasPermission = false
	default:
		hasPermission = false
	}

	// If we got conditional permission but didn't provide context, that's an error
	if resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION &&
		(req.Context == nil || len(req.Context) == 0) {
		return &pb.PermissionResponse{
			HasPermission:  false,
			Permissionship: int32(resp.Permissionship),
			ErrorMessage:   "conditional permission requires context but none was provided",
		}, nil
	}

	return &pb.PermissionResponse{
		HasPermission:  hasPermission,
		Permissionship: int32(resp.Permissionship),
	}, nil
}
