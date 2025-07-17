package config

import (
	"JudgeCore/internal/global"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// 加载JSON配置文件
func LoadConfig() (*global.AppConfig, error) {
	filePath := filepath.Join(global.CurrentDir, "config.json")

	// 检查配置文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果配置文件不存在，创建默认配置
		if err := createDefaultConfig(filePath); err != nil {
			return nil, fmt.Errorf("Failed to create default config file: %v", err)
		}
		log.Println("[FeasOJ] Created default config file:", filePath)
	}

	// 读取配置文件
	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	var config global.AppConfig
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %v", err)
	}

	return &config, nil
}

// 创建默认配置文件
func createDefaultConfig(filePath string) error {
	defaultConfig := global.AppConfig{
		Consul: struct {
			Address     string `json:"address"`
			ServiceName string `json:"service_name"`
			ServiceID   string `json:"service_id"`
		}{
			Address:     "127.0.0.1:8500",
			ServiceName: "JudgeCore",
			ServiceID:   "JudgeCore-1",
		},
		RabbitMQ: struct {
			Address string `json:"address"`
		}{
			Address: "amqp://USER:PASSWORD@IP:PORT/",
		},
		Server: struct {
			Address     string `json:"address"`
			Port        int    `json:"port"`
			EnableHTTPS bool   `json:"enable_https"`
			CertPath    string `json:"cert_path"`
			KeyPath     string `json:"key_path"`
		}{
			Address:     "127.0.0.1",
			Port:        37885,
			EnableHTTPS: false,
			CertPath:    "./certificate/fullchain.pem", // IF USE HTTPS
			KeyPath:     "./certificate/privkey.key",   // IF USE HTTPS
		},
		Sandbox: struct {
			Memory        int64   `json:"memory"`
			NanoCPUs      float64 `json:"nano_cpus"`
			CPUShares     int64   `json:"cpu_shares"`
			MaxConcurrent int     `json:"max_concurrent"`
		}{
			Memory:        2 * 1024 * 1024 * 1024, // 2GB is better, Some compiler will use more memory
			NanoCPUs:      0.5,
			CPUShares:     1024,
			MaxConcurrent: 5,
		},
		Database: struct {
			Address      string `json:"address"`
			Name         string `json:"name"`
			User         string `json:"user"`
			Password     string `json:"password"`
			MaxOpenConns int    `json:"max_open_conns"`
			MaxIdleConns int    `json:"max_idle_conns"`
			MaxLifeTime  int    `json:"max_life_time"`
		}{
			Address:      "IP:PORT",
			Name:         "DATABASE_NAME",
			User:         "DATABASE_USER",
			Password:     "DATABASE_PASSWORD",
			MaxOpenConns: 240,
			MaxIdleConns: 100,
			MaxLifeTime:  32,
		},
	}

	// 将配置写入文件
	configData, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, configData, 0644)
}

// 保存配置到文件
func SaveConfig(config *global.AppConfig) error {
	filePath := filepath.Join(global.CurrentDir, "config.json")
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, configData, 0644)
}

// 初始化配置
func InitConfig() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("[FeasOJ] Failed to load config file: %v", err)
	}

	global.AppConfigInstance = config
	log.Println("[FeasOJ] Config file loaded successfully")
}

// 获取数据库连接字符串
func GetDatabaseDSN() string {
	if global.AppConfigInstance == nil {
		log.Println("[FeasOJ] Config not initialized")
		return ""
	}

	db := global.AppConfigInstance.Database
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		db.User, db.Password, db.Address, db.Name)
}

// 获取Consul地址
func GetConsulAddress() string {
	if global.AppConfigInstance == nil {
		return "localhost:8500"
	}
	return global.AppConfigInstance.Consul.Address
}

// 获取Consul服务名称
func GetConsulServiceName() string {
	if global.AppConfigInstance == nil {
		return "JudgeCore"
	}
	return global.AppConfigInstance.Consul.ServiceName
}

// 获取Consul服务ID
func GetConsulServiceID() string {
	if global.AppConfigInstance == nil {
		return "JudgeCore-1"
	}
	return global.AppConfigInstance.Consul.ServiceID
}

// 获取RabbitMQ地址
func GetRabbitMQAddress() string {
	return global.AppConfigInstance.RabbitMQ.Address
}

// 获取服务器地址
func GetServiceAddress() string {
	if global.AppConfigInstance == nil {
		return "127.0.0.1"
	}
	return global.AppConfigInstance.Server.Address
}

// 获取服务器端口
func GetServicePort() int {
	if global.AppConfigInstance == nil {
		return 37885
	}
	return global.AppConfigInstance.Server.Port
}

// 是否启用HTTPS
func IsHTTPSEnabled() bool {
	if global.AppConfigInstance == nil {
		return false
	}
	return global.AppConfigInstance.Server.EnableHTTPS
}

// 获取服务器证书路径
func GetServerCertPath() string {
	if global.AppConfigInstance == nil {
		return "./certificate/fullchain.pem"
	}
	return global.AppConfigInstance.Server.CertPath
}

// 获取服务器私钥路径
func GetServerKeyPath() string {
	if global.AppConfigInstance == nil {
		return "./certificate/privkey.key"
	}
	return global.AppConfigInstance.Server.KeyPath
}

// 获取沙盒最大并发数
func GetMaxSandbox() int {
	if global.AppConfigInstance == nil {
		return 5
	}
	return global.AppConfigInstance.Sandbox.MaxConcurrent
}

// 获取MySQL最大连接数
func GetMaxOpenConns() int {
	if global.AppConfigInstance == nil {
		return 240
	}
	return global.AppConfigInstance.Database.MaxOpenConns
}

// 获取MySQL最大空闲连接数
func GetMaxIdleConns() int {
	if global.AppConfigInstance == nil {
		return 100
	}
	return global.AppConfigInstance.Database.MaxIdleConns
}

// 获取MySQL连接最大生命周期
func GetMaxLifeTime() int {
	if global.AppConfigInstance == nil {
		return 32
	}
	return global.AppConfigInstance.Database.MaxLifeTime
}

// 获取沙盒内存限制
func GetSandboxMemory() int64 {
	if global.AppConfigInstance == nil {
		return 2 * 1024 * 1024 * 1024 // 2GB
	}
	return global.AppConfigInstance.Sandbox.Memory
}

// 获取沙盒CPU限制
func GetSandboxNanoCPUs() float64 {
	if global.AppConfigInstance == nil {
		return 0.5
	}
	return global.AppConfigInstance.Sandbox.NanoCPUs
}

// 获取沙盒CPU权重
func GetSandboxCPUShares() int64 {
	if global.AppConfigInstance == nil {
		return 1024
	}
	return global.AppConfigInstance.Sandbox.CPUShares
}
