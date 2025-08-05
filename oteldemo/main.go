package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"time"
)

func main() {
	provider := initOTEL()
	defer provider.Shutdown(context.Background())
	otel.SetTracerProvider(provider)

	client := initRedis()
	server := gin.Default()
	// 接入gin的otel插件
	server.Use(otelgin.Middleware("webook"))
	server.GET("/redis", func(ctx *gin.Context) {
		err := client.Set(context.Background(), "aa", "b", time.Second*20).Err()
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"message": "value set"})
	})
	server.Run(":8083")
}

type OTELHook struct {
	tracer trace.Tracer
}

func NewOTELHook() *OTELHook {
	tracer := otel.Tracer("otel_demo/opentelemetry")
	return &OTELHook{
		tracer: tracer,
	}
}

func (o *OTELHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (o *OTELHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	// 接入opentelemetry
	var ctx = context.Background()
	_, span := o.tracer.Start(ctx, "top-span")
	defer span.End()
	span.AddEvent("调用redis")
	fmt.Println("调用redis")
	return func(ctx context.Context, cmd redis.Cmder) error {
		return next(ctx, cmd)
	}
}

func (o *OTELHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "172.31.0.7:6379",
		DB:   0,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("无法连接Redis: %v", err) // 改为Fatal确保问题可见
	}
	client.AddHook(NewOTELHook())
	return client
}
