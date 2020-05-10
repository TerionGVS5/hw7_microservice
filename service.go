package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	//"fmt"
	"google.golang.org/grpc"
	"net"
)

type BizManager struct {
	ACLjson map[string][]string
}

func (b BizManager) ACLService(ctx context.Context, customUrl string) (*Nothing, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &Nothing{}, status.Error(codes.Unauthenticated, "error when load metadata")
	}
	consumerArr := md.Get("consumer")
	consumer := ""
	if consumerArr == nil {
		return &Nothing{}, status.Error(codes.Unauthenticated, "no consumer in ctx")
	} else {
		consumer = consumerArr[0]
	}
	for key, value := range b.ACLjson {
		if key == consumer {
			for _, url := range value {
				if url == customUrl || url == "/main.Biz/*" {
					return &Nothing{}, nil
				}
			}
			return &Nothing{}, status.Error(codes.Unauthenticated, "no url in acl")
		}
	}
	return &Nothing{}, status.Error(codes.Unauthenticated, "no consumer in acl")
}

func (b BizManager) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Check")
}

func (b BizManager) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Add")
}

func (b BizManager) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Test")
}

func NewBizManager(ACLjson map[string][]string) *BizManager {
	return &BizManager{
		ACLjson,
	}
}

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	ACLjson := make(map[string][]string)
	err := json.Unmarshal([]byte(ACLData), &ACLjson)
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	RegisterBizServer(server, NewBizManager(ACLjson))
	go server.Serve(lis)
	go ServerStopper(ctx, server)
	return nil
}

func ServerStopper(ctx context.Context, server *grpc.Server) {
	for {
		select {
		case <-ctx.Done():
			server.Stop()
			return
		}
	}
}
