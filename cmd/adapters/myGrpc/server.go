package myGrpc

import (
	"context"
	"net"

	"github.com/jersonsatoru/cnb/internal/core"
	pb "github.com/jersonsatoru/cnb/internal/pb/proto"
	"google.golang.org/grpc"
)

type Server struct {
	KeyValueStore *core.KeyValueStore
	pb.UnimplementedKeyValueServer
}

func (s *Server) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	value, err := s.KeyValueStore.Get(request.GetKey())
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Value: value}, nil
}

func (s *Server) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutResponse, error) {
	err := s.KeyValueStore.Put(request.Key, request.Value)
	if err != nil {
		return nil, err
	}
	return &pb.PutResponse{}, nil
}

func (s *Server) Delete(ctx context.Context, request *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.KeyValueStore.Delete(request.Key)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteResponse{}, nil
}

func (s *Server) Start(port string) error {
	ss := grpc.NewServer()
	pb.RegisterKeyValueServer(ss, s)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	err = ss.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}
