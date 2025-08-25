package main

import (
	"context"
	"g/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	// 1. 建立与服务器的连接
	conn, err := grpc.Dial("localhost:1234",
		grpc.WithTransportCredentials(insecure.NewCredentials())) // 不安全的连接，仅用于测试
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 2. 创建客户端
	client := service.NewHelloServiceClient(conn)

	// 3. 准备请求
	req := &service.Request{
		Value: "World",
	}

	// 4. 调用RPC方法
	resp, err := client.Hello(context.Background(), req)
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	// 5. 处理响应
	log.Printf("收到响应: %s", resp.Value)
}
