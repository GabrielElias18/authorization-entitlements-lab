package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	graphql "github.com/adityakumar/labs/go-entitlement-service/internal/graphql"
	neo4j "github.com/adityakumar/labs/go-entitlement-service/internal/neo4j"
	pb "github.com/adityakumar/labs/go-entitlement-service/internal/pb/entitlement-service/proto"
	"github.com/adityakumar/labs/go-entitlement-service/internal/service"
	spicedb "github.com/adityakumar/labs/go-entitlement-service/internal/spicedb"
)

type server struct {
	pb.UnimplementedEntitlementServiceServer
	svc *service.Service
}

func (s *server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy:        true,
		Implementation: req.Implementation,
		StatusMessage:  "OK",
	}, nil
}

func (s *server) CheckPermission(ctx context.Context, req *pb.PermissionRequest) (*pb.PermissionResponse, error) {
	return s.svc.CheckPermission(ctx, req)
}

func (s *server) CheckBulkPermissions(ctx context.Context, req *pb.BulkPermissionRequest) (*pb.BulkPermissionResponse, error) {
	if len(req.Requests) == 0 {
		return &pb.BulkPermissionResponse{
			Responses:    []*pb.PermissionResponse{},
			TotalTimeMs:  0,
			SuccessCount: 0,
			ErrorCount:   0,
		}, nil
	}

	startTime := time.Now()
	responses := make([]*pb.PermissionResponse, len(req.Requests))
	successCount := int32(0)
	errorCount := int32(0)

	// Process requests concurrently if max_concurrency is set
	maxConcurrency := req.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = int32(len(req.Requests)) // No limit
	}

	// Use semaphore pattern for concurrency control
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, permReq := range req.Requests {
		wg.Add(1)
		go func(idx int, request *pb.PermissionRequest) {
			defer wg.Done()
			sem <- struct{}{} // Acquire
			defer func() { <-sem }() // Release

			resp, err := s.svc.CheckPermission(ctx, request)
			if err != nil {
				responses[idx] = &pb.PermissionResponse{
					HasPermission: false,
					ErrorMessage:  err.Error(),
				}
				atomic.AddInt32(&errorCount, 1)
			} else {
				responses[idx] = resp
				atomic.AddInt32(&successCount, 1)
			}
		}(i, permReq)
	}

	wg.Wait()
	totalTime := time.Since(startTime).Milliseconds()

	return &pb.BulkPermissionResponse{
		Responses:    responses,
		TotalTimeMs:  float64(totalTime),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	}, nil
}

func (s *server) StreamPermissionChecks(stream pb.EntitlementService_StreamPermissionChecksServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		resp, err := s.svc.CheckPermission(stream.Context(), req)
		if err != nil {
			resp = &pb.PermissionResponse{
				HasPermission: false,
				ErrorMessage:  err.Error(),
			}
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *server) Benchmark(ctx context.Context, req *pb.BenchmarkRequest) (*pb.BenchmarkResponse, error) {
	if len(req.TestCases) == 0 || req.Iterations <= 0 {
		return &pb.BenchmarkResponse{
			Results: []*pb.BenchmarkResult{},
			Summary: &pb.BenchmarkSummary{},
		}, nil
	}

	startTime := time.Now()
	results := make([]*pb.BenchmarkResult, len(req.TestCases))
	var totalRequests int32

	for i, testCase := range req.TestCases {
		result := s.runBenchmarkTest(ctx, testCase, req.Iterations, req.Concurrency, req.Implementation)
		results[i] = result
		totalRequests += result.TotalRequests
	}

	totalTime := time.Since(startTime).Milliseconds()
	avgResponseTime := float64(totalTime) / float64(totalRequests)
	throughput := float64(totalRequests) / (float64(totalTime) / 1000.0) // requests per second

	summary := &pb.BenchmarkSummary{
		TotalTimeMs:       float64(totalTime),
		TotalRequests:     totalRequests,
		AvgResponseTimeMs: avgResponseTime,
		ThroughputRps:     throughput,
		Implementation:    req.Implementation,
	}

	return &pb.BenchmarkResponse{
		Results: results,
		Summary: summary,
	}, nil
}

func (s *server) runBenchmarkTest(ctx context.Context, testCase *pb.TestCase, iterations, concurrency int32, implementation pb.Implementation) *pb.BenchmarkResult {
	if concurrency <= 0 {
		concurrency = 1
	}

	times := make([]float64, iterations)
	successCount := int32(0)
	failedCount := int32(0)

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := int32(0); i < iterations; i++ {
		wg.Add(1)
		go func(idx int32) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			start := time.Now()
			permReq := &pb.PermissionRequest{
				Actor:      testCase.Actor,
				Resource:   testCase.Resource,
				Permission: testCase.Permission,
				Context:    testCase.Context,
			}

			resp, err := s.svc.CheckPermission(ctx, permReq)
			elapsed := time.Since(start).Milliseconds()
			times[idx] = float64(elapsed)

			if err != nil || resp.HasPermission != testCase.ExpectedResult {
				atomic.AddInt32(&failedCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// Calculate statistics
	sort.Float64s(times)
	minTime := times[0]
	maxTime := times[len(times)-1]

	var sum float64
	for _, t := range times {
		sum += t
	}
	avgTime := sum / float64(len(times))

	p95Index := int(float64(len(times)) * 0.95)
	p99Index := int(float64(len(times)) * 0.99)
	p95Time := times[p95Index]
	p99Time := times[p99Index]

	return &pb.BenchmarkResult{
		TestName:           testCase.Name,
		Success:            failedCount == 0,
		AvgResponseTimeMs:  avgTime,
		MinResponseTimeMs:  minTime,
		MaxResponseTimeMs:  maxTime,
		P95ResponseTimeMs:  p95Time,
		P99ResponseTimeMs:  p99Time,
		TotalRequests:      iterations,
		SuccessfulRequests: successCount,
		FailedRequests:     failedCount,
		Implementation:     implementation,
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052" // Use different port from SpiceDB (50051)
	}

	// Initialize backend clients - SpiceDB, Neo4j, and GraphQL
	log.Println("Initializing SpiceDB client...")
	spiceClient := spicedb.NewClient()

	log.Println("Initializing Neo4j client...")
	neo4jClient, err := neo4j.NewClient("bolt://localhost:7687", "neo4j", "password")
	if err != nil {
		log.Printf("Warning: Failed to initialize Neo4j client: %v", err)
		neo4jClient = nil
	} else {
		log.Println("Neo4j client initialized successfully")
	}

	log.Println("Initializing GraphQL client...")
	graphqlClient := graphql.NewClient()
	log.Println("GraphQL client initialized successfully")

	// Initialize service with all backends
	svc := service.NewService(spiceClient, neo4jClient, graphqlClient)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterEntitlementServiceServer(grpcServer, &server{svc: svc})

	// Enable reflection for grpcurl testing
	reflection.Register(grpcServer)

	log.Printf("Entitlement gRPC server listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
