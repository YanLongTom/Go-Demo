package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ESConsumerLogEntry ES消费者使用的日志条目结构
type ESConsumerLogEntry struct {
	Timestamp       string            `json:"@timestamp"`
	Level           string            `json:"level"`
	Service         string            `json:"service"`
	Message         string            `json:"message"`
	FileName        string            `json:"file_name,omitempty"`
	FilePath        string            `json:"file_path,omitempty"`
	LineNumber      int               `json:"line_number,omitempty"`
	RequestID       string            `json:"request_id"`
	OriginalLine    string            `json:"original_line,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CollectedAt     string            `json:"collected_at,omitempty"`
	ProcessedAt     string            `json:"processed_at,omitempty"`
	CleanedMessage  string            `json:"cleaned_message,omitempty"`
	ExtractedFields map[string]string `json:"extracted_fields,omitempty"`
	DataQuality     string            `json:"data_quality,omitempty"`

	// 防重复字段
	MessageHash  string `json:"message_hash"`
	EsDocumentID string `json:"es_document_id"`
	ConsumedAt   string `json:"consumed_at"`
	RetryCount   int    `json:"retry_count"`
}

// DuplicateChecker 重复检查器
type DuplicateChecker struct {
	processedMessages map[string]bool
	mutex             sync.RWMutex
	maxSize           int
}

// NewDuplicateChecker 创建重复检查器
func NewDuplicateChecker(maxSize int) *DuplicateChecker {
	return &DuplicateChecker{
		processedMessages: make(map[string]bool),
		maxSize:           maxSize,
	}
}

// IsProcessed 检查是否已处理
func (dc *DuplicateChecker) IsProcessed(messageID string) bool {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.processedMessages[messageID]
}

// MarkProcessed 标记为已处理
func (dc *DuplicateChecker) MarkProcessed(messageID string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	// 如果超过最大大小，清理一半
	if len(dc.processedMessages) >= dc.maxSize {
		newMap := make(map[string]bool)
		count := 0
		for k, v := range dc.processedMessages {
			if count < dc.maxSize/2 {
				newMap[k] = v
				count++
			}
		}
		dc.processedMessages = newMap
		log.Printf("🧹 重复检查器清理完成，保留 %d 条记录", count)
	}

	dc.processedMessages[messageID] = true
}

// ElasticsearchClient Elasticsearch客户端
type ElasticsearchClient struct {
	client           *elasticsearch.Client
	duplicateChecker *DuplicateChecker
}

// NewElasticsearchClient 创建Elasticsearch客户端
func NewElasticsearchClient(addresses []string) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		// 优化连接配置
		MaxRetries:    5,
		RetryOnStatus: []int{502, 503, 504, 429},
		// 连接池配置
		DiscoverNodesOnStart:  true,
		DiscoverNodesInterval: 60 * time.Second,
		// 连接超时
		//		DialTimeout:    30 * time.Second,
		//		RequestTimeout: 30 * time.Second,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建ES客户端失败: %v", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("ES连接测试失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES服务不可用: %s", res.String())
	}

	log.Println("✅ Elasticsearch连接成功")
	return &ElasticsearchClient{
		client:           client,
		duplicateChecker: NewDuplicateChecker(10000), // 缓存1万条记录
	}, nil
}

// generateDocumentID 生成文档ID（防重复的关键）
func (es *ElasticsearchClient) generateDocumentID(logEntry *ESConsumerLogEntry) string {
	// 方案1：使用RequestID（推荐）
	if logEntry.RequestID != "" {
		return logEntry.RequestID
	}

	// 方案2：使用内容哈希
	content := fmt.Sprintf("%s_%s_%s_%s_%d",
		logEntry.Timestamp, logEntry.Service, logEntry.Level, logEntry.Message, logEntry.LineNumber)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16] // 取前16位
}

// IndexDocument 索引文档到Elasticsearch (重点优化防重复)
func (es *ElasticsearchClient) IndexDocument(index string, logEntry *ESConsumerLogEntry) error {
	// 生成唯一文档ID
	docID := es.generateDocumentID(logEntry)
	logEntry.EsDocumentID = docID

	// 生成消息哈希用于重复检查
	messageContent := fmt.Sprintf("%s_%s_%s", logEntry.RequestID, logEntry.Timestamp, logEntry.Message)
	messageHash := sha256.Sum256([]byte(messageContent))
	logEntry.MessageHash = hex.EncodeToString(messageHash[:])[:16]

	// 检查是否已处理过
	if es.duplicateChecker.IsProcessed(logEntry.MessageHash) {
		log.Printf("⚠️  重复消息，跳过: ID=%s", docID)
		return nil
	}

	// 添加时间戳
	logEntry.ConsumedAt = time.Now().Format(time.RFC3339)

	jsonDoc, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 🔑 关键配置：防重复写入
	req := esapi.IndexRequest{
		Index:         index,
		DocumentID:    docID, // 使用确定性ID
		Body:          strings.NewReader(string(jsonDoc)),
		Refresh:       "wait_for", // 确保立即可查询
		OpType:        "create",   // 🚨 重要：只创建，不覆盖
		Timeout:       time.Second * 30,
		IfSeqNo:       nil,
		IfPrimaryTerm: nil,
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("ES索引请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 如果是冲突错误（文档已存在），则认为成功
		if res.StatusCode == 409 {
			log.Printf("📝 文档已存在，跳过: ID=%s", docID)
			es.duplicateChecker.MarkProcessed(logEntry.MessageHash)
			return nil
		}

		// 其他错误，解析详情
		var errResponse map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&errResponse); err == nil {
			return fmt.Errorf("ES索引失败 [%s]: %v", res.Status(), errResponse)
		}
		return fmt.Errorf("ES索引失败: %s", res.String())
	}

	// 标记为已处理
	es.duplicateChecker.MarkProcessed(logEntry.MessageHash)

	log.Printf("✅ 文档索引成功: ID=%s, Hash=%s", docID, logEntry.MessageHash[:8])
	return nil
}

// CreateIndexTemplate 创建索引模板 (重要：优化ES性能)
func (es *ElasticsearchClient) CreateIndexTemplate() error {
	template := `{
		"index_patterns": ["logs-*"],
		"template": {
			"settings": {
				"number_of_shards": 1,
				"number_of_replicas": 0,
				"refresh_interval": "5s",
				"index": {
					"max_result_window": 50000,
					"mapping": {
						"ignore_malformed": true
					}
				}
			},
			"mappings": {
				"properties": {
					"@timestamp": {
						"type": "date",
						"format": "strict_date_optional_time||epoch_millis"
					},
					"level": {
						"type": "keyword"
					},
					"service": {
						"type": "keyword"
					},
					"message": {
						"type": "text",
						"analyzer": "standard"
					},
					"cleaned_message": {
						"type": "text",
						"analyzer": "standard"
					},
					"request_id": {
						"type": "keyword"
					},
					"file_name": {
						"type": "keyword"
					},
					"line_number": {
						"type": "integer"
					},
					"data_quality": {
						"type": "keyword"
					},
					"message_hash": {
						"type": "keyword"
					},
					"es_document_id": {
						"type": "keyword"
					},
					"retry_count": {
						"type": "integer"
					}
				}
			}
		}
	}`

	req := esapi.IndicesPutIndexTemplateRequest{
		Name: "logs-template",
		Body: strings.NewReader(template),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("创建索引模板失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("警告: 创建索引模板失败: %s", res.String())
	} else {
		log.Println("✅ 索引模板创建成功")
	}

	return nil
}

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topics   []string
	groupID  string
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(brokers []string, groupID string, topics []string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 20 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 6 * time.Second
	config.Consumer.MaxProcessingTime = 10 * time.Second
	config.Consumer.Return.Errors = true

	// 🔑 关键配置：防止重复消费
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	config.Consumer.Group.Rebalance.Timeout = 60 * time.Second

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %v", err)
	}

	log.Printf("✅ Kafka消费者创建成功，消费组: %s", groupID)
	return &KafkaConsumer{
		consumer: consumer,
		topics:   topics,
		groupID:  groupID,
	}, nil
}

// ConsumerGroupHandler 消费者组处理器
type ConsumerGroupHandler struct {
	esClient *ElasticsearchClient
	ready    chan bool
}

// Setup 消费者组设置
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup 消费者组清理
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费消息 (重点优化：ES写入逻辑)
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			log.Printf("📥 接收消息: topic=%s, partition=%d, offset=%d, size=%d bytes",
				message.Topic, message.Partition, message.Offset, len(message.Value))

			// 解析日志数据
			var logEntry ESConsumerLogEntry
			if err := json.Unmarshal(message.Value, &logEntry); err != nil {
				log.Printf("❌ JSON解析失败: %v, 原始数据: %s", err, string(message.Value))
				session.MarkMessage(message, "")
				continue
			}

			// 确保RequestID存在
			if logEntry.RequestID == "" {
				logEntry.RequestID = fmt.Sprintf("kafka_%s_%d_%d", message.Topic, message.Partition, message.Offset)
			}

			// 🔑 确保@timestamp字段存在并格式正确
			if logEntry.Timestamp == "" {
				logEntry.Timestamp = time.Now().Format(time.RFC3339)
			}

			// 生成索引名（按日期分片）
			indexName := fmt.Sprintf("logs-%s", time.Now().Format("2006-01-02"))

			// 🔑 ES写入重试机制（最大3次）
			maxRetries := 3
			var indexErr error

			for retry := 0; retry < maxRetries; retry++ {
				logEntry.RetryCount = retry
				indexErr = h.esClient.IndexDocument(indexName, &logEntry)

				if indexErr == nil {
					break // 成功，跳出重试循环
				}

				log.Printf("⚠️  ES写入失败 (重试 %d/%d): %v", retry+1, maxRetries, indexErr)

				if retry < maxRetries-1 {
					// 递增延迟重试：1s, 2s, 3s
					time.Sleep(time.Duration(retry+1) * time.Second)
				}
			}

			if indexErr != nil {
				log.Printf("❌ ES写入最终失败，跳过消息: %v", indexErr)
				// 失败的消息也要标记，避免无限重试导致消费阻塞
				session.MarkMessage(message, "")
				continue
			}

			log.Printf("✅ 日志索引成功: Service=%s, Level=%s, Quality=%s",
				logEntry.Service, logEntry.Level, logEntry.DataQuality)

			// 🔑 重要：只有成功写入ES后才提交offset
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// StartConsuming 开始消费
func (kc *KafkaConsumer) StartConsuming(esClient *ElasticsearchClient) error {
	handler := &ConsumerGroupHandler{
		esClient: esClient,
		ready:    make(chan bool),
	}

	ctx := context.Background()
	go func() {
		for {
			if err := kc.consumer.Consume(ctx, kc.topics, handler); err != nil {
				log.Printf("❌ 消费错误: %v", err)
			}
			// 检查上下文是否被取消
			if ctx.Err() != nil {
				return
			}
			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready
	log.Println("🚀 Kafka消费者已准备就绪")

	// 处理错误
	go func() {
		for err := range kc.consumer.Errors() {
			log.Printf("❌ 消费者错误: %v", err)
		}
	}()

	return nil
}

// Close 关闭消费者
func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}

func main() {
	log.Println("========================================")
	log.Println("🚀 启动Elasticsearch消费者")
	log.Println("🔒 防重复消费 & 防重复写入优化版")
	log.Println("========================================")

	// 创建Elasticsearch客户端
	esClient, err := NewElasticsearchClient([]string{"http://localhost:9200"})
	if err != nil {
		log.Fatalf("❌ 创建Elasticsearch客户端失败: %v", err)
	}

	// 创建索引模板（重要：优化ES性能）
	err = esClient.CreateIndexTemplate()
	if err != nil {
		log.Printf("⚠️  创建索引模板失败: %v", err)
	}

	// 创建Kafka消费者
	consumer, err := NewKafkaConsumer(
		[]string{"localhost:9092"},
		"log-consumer-group-v2", // 使用新的消费组
		[]string{"log-data"},
	)
	if err != nil {
		log.Fatalf("❌ 创建Kafka消费者失败: %v", err)
	}
	defer consumer.Close()

	log.Println("📊 配置信息:")
	log.Println("  Kafka Brokers: localhost:9092")
	log.Println("  Consumer Group: log-consumer-group-v2")
	log.Println("  Topic: log-data")
	log.Println("  Elasticsearch: http://localhost:9200")
	log.Println("  防重复策略: 文档ID + 内容哈希")
	log.Println("  重试机制: 最大3次，递增延迟")
	log.Println("========================================")

	// 开始消费
	err = consumer.StartConsuming(esClient)
	if err != nil {
		log.Fatalf("❌ 开始消费失败: %v", err)
	}

	log.Println("🎯 开始消费日志数据，按 Ctrl+C 停止...")

	// 保持程序运行
	select {}
}
