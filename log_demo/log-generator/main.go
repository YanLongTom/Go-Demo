package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// LogFileConfig 日志文件配置
type LogFileConfig struct {
	FileName string
	Service  string
	LogCount int
}

// generateLogLine 生成单行日志
func generateLogLine(service string) string {
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	messages := map[string][]string{
		"nginx": {
			"GET /api/users HTTP/1.1 200 0.123",
			"POST /api/login HTTP/1.1 401 0.056",
			"GET /static/css/style.css HTTP/1.1 404 0.012",
			"POST /api/orders HTTP/1.1 500 1.234",
			"GET /health HTTP/1.1 200 0.001",
			"PUT /api/users/123 HTTP/1.1 200 0.445",
		},
		"application": {
			"用户登录成功 userId=12345",
			"订单创建失败 orderId=98765 error=库存不足",
			"支付处理完成 paymentId=54321 amount=199.99",
			"数据库连接超时 connection=pool-1",
			"缓存更新成功 key=user:12345",
			"API调用失败 service=payment-service timeout=5000ms",
		},
		"database": {
			"查询执行完成 table=users duration=12ms rows=156",
			"慢查询警告 table=orders duration=3456ms",
			"连接池满载 active=100 max=100",
			"索引重建开始 table=products",
			"备份任务完成 size=2.3GB duration=45min",
			"死锁检测 table=inventory session=1234",
		},
		"security": {
			"登录尝试 ip=192.168.1.100 user=admin result=success",
			"可疑登录 ip=123.45.67.89 user=admin result=blocked",
			"权限验证失败 user=guest resource=/admin/dashboard",
			"密码重置请求 user=john@example.com",
			"API密钥过期 key=app-12345 service=mobile-app",
			"防火墙阻止 ip=10.0.0.50 port=22 protocol=ssh",
		},
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	level := levels[rand.Intn(len(levels))]
	message := messages[service][rand.Intn(len(messages[service]))]
	requestId := fmt.Sprintf("req_%d", rand.Intn(100000))

	return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, requestId, message)
}

// createLogFile 创建日志文件
func createLogFile(config LogFileConfig, logDir string) error {
	filePath := filepath.Join(logDir, config.FileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %v", filePath, err)
	}
	defer file.Close()

	log.Printf("正在生成 %s (%d 行日志)...", config.FileName, config.LogCount)

	for i := 0; i < config.LogCount; i++ {
		logLine := generateLogLine(config.Service)
		_, err := file.WriteString(logLine + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		// 模拟实时日志生成，每100毫秒生成一条
		time.Sleep(time.Millisecond * 100)
	}

	log.Printf("✅ %s 生成完成", config.FileName)
	return nil
}

// appendLogFile 追加日志到现有文件
func appendLogFile(config LogFileConfig, logDir string) error {
	filePath := filepath.Join(logDir, config.FileName)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败 %s: %v", filePath, err)
	}
	defer file.Close()

	for i := 0; i < config.LogCount; i++ {
		logLine := generateLogLine(config.Service)
		_, err := file.WriteString(logLine + "\n")
		if err != nil {
			return fmt.Errorf("追加文件失败: %v", err)
		}

		// 模拟持续日志生成
		time.Sleep(time.Second * 2)
	}

	return nil
}

func main() {
	// 创建日志目录
	logDir := "./logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 日志文件配置
	logConfigs := []LogFileConfig{
		{FileName: "nginx.log", Service: "nginx", LogCount: 50},
		{FileName: "application.log", Service: "application", LogCount: 30},
		{FileName: "database.log", Service: "database", LogCount: 25},
		{FileName: "security.log", Service: "security", LogCount: 20},
	}

	log.Println("========================================")
	log.Println("开始生成日志文件...")
	log.Println("========================================")

	// 初始生成日志文件
	for _, config := range logConfigs {
		err := createLogFile(config, logDir)
		if err != nil {
			log.Printf("生成日志文件失败: %v", err)
			continue
		}
	}

	log.Println("========================================")
	log.Println("初始日志文件生成完成！")
	log.Println("========================================")
	log.Printf("日志文件位置: %s", logDir)
	log.Println("文件列表:")

	// 显示生成的文件
	files, err := os.ReadDir(logDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				info, err := file.Info()
				if err == nil {
					log.Printf("  📄 %s (%.2f KB)", file.Name(), float64(info.Size())/1024)
				}
			}
		}
	}

	// 询问是否继续生成
	log.Println("\n是否继续生成新日志? (y/n):")
	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" || input == "yes" {
		log.Println("开始持续生成日志 (按 Ctrl+C 停止)...")

		// 持续生成日志
		for {
			for _, config := range logConfigs {
				// 每次只追加少量日志
				config.LogCount = rand.Intn(5) + 1
				go func(cfg LogFileConfig) {
					err := appendLogFile(cfg, logDir)
					if err != nil {
						log.Printf("追加日志失败: %v", err)
					}
				}(config)
			}

			// 每10秒生成一轮新日志
			time.Sleep(time.Second * 10)
		}
	}

	log.Println("日志生成器结束")
}
