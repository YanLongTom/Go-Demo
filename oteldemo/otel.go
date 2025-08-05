package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"time"

	"go.opentelemetry.io/otel/exporters/zipkin"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// initOTEL 初始化OpenTelemetry追踪系统
// 包含三个主要步骤：
// 1. 创建Resource - 标识服务的元数据
// 2. 创建Propagator - 设置上下文传播方式
// 3. 创建TraceProvider - 配置追踪提供者
func initOTEL() *trace.TracerProvider {
	// 1. 创建Resource
	res, err := newResource("demo", "v0.0.1")
	if err != nil {
		panic(err)
	}
	// 2. 创建propagator
	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)
	// 3. 创建traceProvider
	provider := newTraceProvider(res)
	return provider

}

// newResource 创建OpenTelemetry资源
// 参数:	serviceName: 服务名称	serviceVersion: 服务版本
// 返回值:	*resource.Resource: OpenTelemetry资源对象	error: 错误信息
func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
}

// newPropagator 创建文本映射传播器
// 返回值:	propagation.TextMapPropagator: 组合传播器，包含TraceContext和Baggage
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider 创建追踪提供者
// 参数:	res: OpenTelemetry资源对象
// 返回值:	*trace.TracerProvider: 追踪提供者实例
func newTraceProvider(res *resource.Resource) *trace.TracerProvider {
	exporter, err := zipkin.New("http://172.31.0.7:9411/api/v2/spans")
	if err != nil {
		panic(err)
	}
	return trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
}
