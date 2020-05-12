package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"strings"
	"sync"
	"time"

	//"fmt"
	"google.golang.org/grpc"
	"net"
)

type BaseManager struct {
	ACLjson map[string][]string
	mu      *sync.RWMutex
	events  map[int]*Event
	host    string
}

type BizManager struct {
	BaseManager
}

type AdmManager struct {
	BaseManager
}

func contain(arr []int, number int) bool {
	for _, el := range arr {
		if el == number {
			return true
		}
	}
	return false
}

func (a AdmManager) Logging(n *Nothing, logServer Admin_LoggingServer) error {
	_, err := a.ACLService(logServer.Context(), "/main.Admin/Logging", "/main.Admin/*")
	started := int(time.Now().UnixNano())
	lastTimeSend := time.Now().Unix()
	if err != nil {
		return err
	}
	sendIndexes := make([]int, 0)
	for {
		a.mu.RLock()
		for key, value := range a.events {
			if key > started && !contain(sendIndexes, key) {
				sendIndexes = append(sendIndexes, key)
				logServer.Send(value)
				lastTimeSend = time.Now().Unix()
			}
			if lastTimeSend+3 <= time.Now().Unix() {
				return nil
			}
		}
		a.mu.RUnlock()
	}
	return nil
}

func (a AdmManager) Statistics(interval *StatInterval, statServer Admin_StatisticsServer) error {
	_, err := a.ACLService(statServer.Context(), "/main.Admin/Statistics", "/main.Admin/*")
	started := int(time.Now().UnixNano())
	if err != nil {
		return err
	}
	sendIndexes := make([]int, 0)
	for {
		wasNewStat := false
		time.Sleep(time.Second * time.Duration(interval.IntervalSeconds))
		newStat := Stat{
			ByConsumer: make(map[string]uint64),
			ByMethod:   make(map[string]uint64),
		}
		a.mu.RLock()
		for key, value := range a.events {
			if key > started && !contain(sendIndexes, key) {
				sendIndexes = append(sendIndexes, key)
				wasNewStat = true
				if countConsumer, ok := newStat.ByConsumer[value.Consumer]; ok {
					newStat.ByConsumer[value.Consumer] = countConsumer + 1
				} else {
					newStat.ByConsumer[value.Consumer] = 1
				}
				if countMethod, ok := newStat.ByMethod[value.Method]; ok {
					newStat.ByMethod[value.Method] = countMethod + 1
				} else {
					newStat.ByMethod[value.Method] = 1
				}
			}
		}
		a.mu.RUnlock()
		if !wasNewStat {
			return io.EOF
		} else {
			statServer.Send(&newStat)
		}
	}
	return nil
}

func (b BaseManager) ACLService(ctx context.Context, customUrl string, allAccessUrl string) (*Nothing, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &Nothing{}, status.Error(codes.Unauthenticated, "error when load metadata")
	}
	consumerArr := md.Get("consumer")
	consumer := ""
	nowNano := time.Now().UnixNano()
	if consumerArr == nil {
		return &Nothing{}, status.Error(codes.Unauthenticated, "no consumer in ctx")
	} else {
		consumer = consumerArr[0]
	}
	for key, value := range b.ACLjson {
		if key == consumer {
			for _, url := range value {
				if url == customUrl || url == allAccessUrl {
					b.events[int(nowNano)] = &Event{Method: customUrl, Consumer: consumer, Timestamp: nowNano, Host: b.host}
					return &Nothing{}, nil
				}
			}
			return &Nothing{}, status.Error(codes.Unauthenticated, "no url in acl")
		}
	}
	return &Nothing{}, status.Error(codes.Unauthenticated, "no consumer in acl")
}

func (b BizManager) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Check", "/main.Biz/*")
}

func (b BizManager) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Add", "/main.Biz/*")
}

func (b BizManager) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return b.ACLService(ctx, "/main.Biz/Test", "/main.Biz/*")
}

func NewBaseManager(ACLjson map[string][]string, listenAddr string) *BaseManager {
	return &BaseManager{
		ACLjson: ACLjson,
		mu:      &sync.RWMutex{},
		events:  make(map[int]*Event),
		host:    fmt.Sprintf(`%s:`, strings.Split(listenAddr, ":")[0]),
	}
}

func NewBizManager(baseManager BaseManager) *BizManager {
	return &BizManager{baseManager}
}

func NewAdmManager(baseManager BaseManager) *AdmManager {
	return &AdmManager{baseManager}
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
	baseManager := NewBaseManager(ACLjson, listenAddr)
	RegisterBizServer(server, NewBizManager(*baseManager))
	RegisterAdminServer(server, NewAdmManager(*baseManager))
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
