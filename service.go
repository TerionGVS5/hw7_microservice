package main

import (
	"context"
	//"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listet port", err)
		return err
	}

	server := grpc.NewServer()
	//fmt.Println(fmt.Sprintf(`before starting server at: %s`, listenAddr))
	go server.Serve(lis)
	go ServerStopper(ctx, server)
	//fmt.Println(fmt.Sprintf(`after starting server at: %s`, listenAddr))

	return nil
}

func ServerStopper(ctx context.Context, server *grpc.Server) {
	for {
		select {
		case <- ctx.Done():
			server.Stop()
			return
		}
	}
}