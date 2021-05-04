package main

import (
	"sync"

	"google.golang.org/grpc"
	pb "github.com/stevealexrs/Go-Libra/service/public"
)

type publicServer struct {
	pb.Un
	savedFeatures []*pb.Feature // read-only after initialized

	mu sync.Mutex // protects routeNotes
}