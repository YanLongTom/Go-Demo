package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/fsnotify/fsnotify"
)

// FileCollectorLogEntry 文件采集器日志条目
type FileCollectorLogEntry struct {
	Timestamp    string            `json:"timestamp"`
	Level        string            `json:"level"`
	Service      string            `json:"service"`
	Message      string            `json:"message"`
	FileName     string            `json:"file_name"`
	FilePath     string            `json:"file_path"`
	LineNumber   int               `json:"line_number"`
	RequestID    string            `json:"request_id"`
	OriginalLine string            `json:"original_line"`
	Metadata     map[string]string `json:"metadata"`
	CollectedAt  string            `json:"collected_at"`

	// 清洗后的数据
	CleanedMessage  string            `json:"cleaned_message"`
	ExtractedFields map[string]string `json:"extracted_fields"`
	ProcessedAt     string            `json:"processed_at"`
	DataQuality     string            `json:"data_quality"`
}

// PythonDataCleaner 真实的Python数据清洗器
type PythonDataCleaner struct {
	pythonScript string
}

// NewPythonDataCleaner 创建数据清洗器
func NewPythonDataCleaner() *PythonDataCleaner {
	// 获取当前目录下的Python脚本路径
	scriptPath := "./data_cleaner.py"
	return &PythonDataCleaner{
		pythonScript: scriptPath,
	}
}

// CleanData 调用Python脚本进行数据清洗
func (cleaner *PythonDataCleaner) CleanData(logEntry *FileCollectorLogEntry) {
	log.Printf("🐍 [调用Python] 处理日志: %s", logEntry.RequestID)

	// 将Go结构体转换为JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("❌ JSON序列化失败: %v", err)
		return
	}

	// 调用Python脚本
	result, err := cleaner.callPythonScript(string(jsonData))
	if err != nil {
		log.Printf("❌ Python脚本调用失败: %v", err)
		// 设置默认值，防止处理失败
		logEntry.CleanedMessage = logEntry.Message
		logEntry.ExtractedFields = make(map[string]string)
		logEntry.ProcessedAt = time.Now().Format(time.RFC3339)
		logEntry.DataQuality = "LOW"
		return
	}

	// 解析Python返回的结果
	var cleanedEntry FileCollectorLogEntry
	if err := json.Unmarshal([]byte(result), &cleanedEntry); err != nil {
		log.Printf("❌ Python结果解析失败: %v", err)
		return
	}

	// 更新原始日志条目
	logEntry.CleanedMessage = cleanedEntry.CleanedMessage
	logEntry.ExtractedFields = cleanedEntry.ExtractedFields
	logEntry.ProcessedAt = cleanedEntry.ProcessedAt
	logEntry.DataQuality = cleanedEntry.DataQuality

	log.Printf("✅ [Python处理] 完成: %s -> %s [质量:%s]",
		logEntry.Message[:min(30, len(logEntry.Message))],
		logEntry.CleanedMessage[:min(30, len(logEntry.CleanedMessage))],
		logEntry.DataQuality)
}

// callPythonScript 调用Python脚本
func (cleaner *PythonDataCleaner) callPythonScript(jsonData string) (string, error) {
	// 构建Python命令
	cmd := exec.Command("python3", cleaner.pythonScript, jsonData)

	// 如果python3不存在，尝试python
	if _, err := exec.LookPath("python3"); err != nil {
		cmd = exec.Command("python", cleaner.pythonScript, jsonData)
	}

	// 执行命令并获取输出
	output, err := cmd.Output()
	if err != nil {
		// 如果命令执行失败，尝试获取错误信息
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("Python脚本执行失败: %s, 错误输出: %s", err, string(exitError.Stderr))
		}
		return "", fmt.Errorf("Python脚本执行失败: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// FileCollectorKafkaProducer 文件采集器专用的Kafka生产者
type FileCollectorKafkaProducer struct {
	producer sarama.SyncProducer
}

// NewFileCollectorKafkaProducer 创建文件采集器Kafka生产者
func NewFileCollectorKafkaProducer(brokers []string) (*FileCollectorKafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有副本确认
	config.Producer.Retry.Max = 5                    // 重试次数
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Idempotent = true // 幂等性，防止重复发送
	config.Net.MaxOpenRequests = 1    // 🔑 幂等性生产者必需配置

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka生产者失败: %v", err)
	}

	return &FileCollectorKafkaProducer{producer: producer}, nil
}

// SendMessage 发送消息到Kafka
func (kp *FileCollectorKafkaProducer) SendMessage(topic string, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("📤 消息发送成功: topic=%s, partition=%d, offset=%d", topic, partition, offset)
	return nil
}

// Close 关闭生产者
func (kp *FileCollectorKafkaProducer) Close() error {
	return kp.producer.Close()
}

// FileCollector 文件采集器
type FileCollector struct {
	kafka         *FileCollectorKafkaProducer
	watcher       *fsnotify.Watcher
	logDir        string
	topic         string
	filePositions map[string]int64
	dataCleaner   *PythonDataCleaner
}

// LogParser 日志解析器
type LogParser struct {
	logPattern *regexp.Regexp
	serviceMap map[string]string
}

// NewLogParser 创建日志解析器
func NewLogParser() *LogParser {
	// 匹配日志格式: [timestamp] [level] [request_id] message
	pattern := regexp.MustCompile(`\[([^\]]+)\]\s*\[([^\]]+)\]\s*\[([^\]]+)\]\s*(.+)`)

	serviceMap := map[string]string{
		"nginx.log":       "nginx",
		"application.log": "application",
		"database.log":    "database",
		"security.log":    "security",
	}

	return &LogParser{
		logPattern: pattern,
		serviceMap: serviceMap,
	}
}

// ParseLogLine 解析日志行
func (p *LogParser) ParseLogLine(line, fileName string, lineNumber int) (*FileCollectorLogEntry, error) {
	matches := p.logPattern.FindStringSubmatch(line)
	if len(matches) != 5 {
		// 如果不匹配标准格式，作为普通消息处理
		return &FileCollectorLogEntry{
			Timestamp:    time.Now().Format(time.RFC3339),
			Level:        "INFO",
			Service:      p.getServiceFromFileName(fileName),
			Message:      strings.TrimSpace(line),
			FileName:     fileName,
			LineNumber:   lineNumber,
			RequestID:    fmt.Sprintf("file_%d_%d", time.Now().Unix(), lineNumber),
			OriginalLine: line,
			CollectedAt:  time.Now().Format(time.RFC3339),
			Metadata:     make(map[string]string),
		}, nil
	}

	// 解析时间戳
	timestamp := matches[1]
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
		timestamp = parsedTime.Format(time.RFC3339)
	}

	// 提取元数据
	metadata := p.extractMetadata(matches[4])

	return &FileCollectorLogEntry{
		Timestamp:    timestamp,
		Level:        matches[2],
		Service:      p.getServiceFromFileName(fileName),
		Message:      matches[4],
		FileName:     fileName,
		LineNumber:   lineNumber,
		RequestID:    matches[3],
		OriginalLine: line,
		CollectedAt:  time.Now().Format(time.RFC3339),
		Metadata:     metadata,
	}, nil
}

// getServiceFromFileName 从文件名获取服务名
func (p *LogParser) getServiceFromFileName(fileName string) string {
	if service, exists := p.serviceMap[fileName]; exists {
		return service
	}
	return "unknown"
}

// extractMetadata 提取元数据
func (p *LogParser) extractMetadata(message string) map[string]string {
	metadata := make(map[string]string)

	// 提取键值对 (key=value 格式)
	kvPattern := regexp.MustCompile(`(\w+)=([^\s]+)`)
	matches := kvPattern.FindAllStringSubmatch(message, -1)

	for _, match := range matches {
		if len(match) == 3 {
			metadata[match[1]] = match[2]
		}
	}

	return metadata
}

// NewFileCollector 创建文件采集器
func NewFileCollector(logDir, topic string, kafka *FileCollectorKafkaProducer) (*FileCollector, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileCollector{
		kafka:         kafka,
		watcher:       watcher,
		logDir:        logDir,
		topic:         topic,
		filePositions: make(map[string]int64),
		dataCleaner:   NewPythonDataCleaner(),
	}, nil
}

// Start 启动文件采集
func (fc *FileCollector) Start() error {
	parser := NewLogParser()

	// 添加目录监控
	err := fc.watcher.Add(fc.logDir)
	if err != nil {
		return err
	}

	log.Printf("开始监控目录: %s", fc.logDir)

	// 首次读取所有现有文件
	err = fc.readExistingFiles(parser)
	if err != nil {
		log.Printf("读取现有文件失败: %v", err)
	}

	// 监控文件变化
	go func() {
		for {
			select {
			case event, ok := <-fc.watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("文件变化: %s", event.Name)
					fc.processFileChange(event.Name, parser)
				}

			case err, ok := <-fc.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("监控错误: %v", err)
			}
		}
	}()

	return nil
}

// readExistingFiles 读取现有文件
func (fc *FileCollector) readExistingFiles(parser *LogParser) error {
	files, err := os.ReadDir(fc.logDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		filePath := filepath.Join(fc.logDir, file.Name())
		log.Printf("处理现有文件: %s", filePath)

		err := fc.processFile(filePath, parser)
		if err != nil {
			log.Printf("处理文件失败 %s: %v", filePath, err)
		}
	}

	return nil
}

// processFileChange 处理文件变化
func (fc *FileCollector) processFileChange(filePath string, parser *LogParser) {
	if !strings.HasSuffix(filePath, ".log") {
		return
	}

	err := fc.processFile(filePath, parser)
	if err != nil {
		log.Printf("处理文件变化失败 %s: %v", filePath, err)
	}
}

// processFile 处理单个文件
func (fc *FileCollector) processFile(filePath string, parser *LogParser) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件当前位置
	fileName := filepath.Base(filePath)
	lastPosition := fc.filePositions[filePath]

	// 定位到上次读取的位置
	_, err = file.Seek(lastPosition, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	lineNumber := int(lastPosition) // 简化的行号计算

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineNumber++

		// 解析日志行
		logEntry, err := parser.ParseLogLine(line, fileName, lineNumber)
		if err != nil {
			log.Printf("解析日志行失败: %v", err)
			continue
		}

		logEntry.FilePath = filePath

		// 🐍 数据清洗处理
		fc.dataCleaner.CleanData(logEntry)

		// 发送到Kafka
		err = fc.sendToKafka(logEntry)
		if err != nil {
			log.Printf("发送到Kafka失败: %v", err)
			continue
		}

		log.Printf("采集日志: %s [%s] %s [质量:%s]", logEntry.Service, logEntry.Level, logEntry.CleanedMessage[:min(40, len(logEntry.CleanedMessage))], logEntry.DataQuality)
	}

	// 更新文件位置
	currentPosition, _ := file.Seek(0, 1)
	fc.filePositions[filePath] = currentPosition

	return scanner.Err()
}

// sendToKafka 发送到Kafka
func (fc *FileCollector) sendToKafka(logEntry *FileCollectorLogEntry) error {
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	return fc.kafka.SendMessage(fc.topic, logEntry.RequestID, jsonData)
}

// Close 关闭采集器
func (fc *FileCollector) Close() error {
	return fc.watcher.Close()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// 创建Kafka生产者
	producer, err := NewFileCollectorKafkaProducer([]string{"localhost:9092"})
	if err != nil {
		log.Fatalf("❌ 创建Kafka生产者失败: %v", err)
	}
	defer producer.Close()

	// 创建文件采集器
	logDir := "/root/A-log/log-generator/logs"
	collector, err := NewFileCollector(logDir, "log-data", producer)
	if err != nil {
		log.Fatalf("❌ 创建文件采集器失败: %v", err)
	}
	defer collector.Close()

	log.Println("========================================")
	log.Println("📂 文件日志采集器启动")
	log.Println("🐍 真实Python数据清洗脚本调用")
	log.Println("========================================")
	log.Printf("📁 监控目录: %s", logDir)
	log.Printf("📤 Kafka主题: log-data")
	log.Printf("📋 支持文件: *.log")
	log.Printf("🐍 Python脚本: ./data_cleaner.py")
	log.Printf("🧹 数据清洗: 真实Python调用")
	log.Println("========================================")

	// 启动采集
	err = collector.Start()
	if err != nil {
		log.Fatalf("❌ 启动采集失败: %v", err)
	}

	log.Println("🚀 文件采集器已启动，按 Ctrl+C 停止...")

	// 保持程序运行
	select {}
}
