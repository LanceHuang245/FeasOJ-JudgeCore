package main

import (
	"bufio"
	"fmt"
	"log"
	"main/internal/config"
	"main/internal/global"
	"main/internal/judge"
	"main/internal/router"
	"main/internal/utils"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
)

func main() {
	global.CurrentDir, _ = os.Getwd()
	global.ParentDir = filepath.Dir(global.CurrentDir)

	// 定义目录映射
	dirs := map[string]*string{
		"certificate": &global.CertDir,
		"configs":     &global.ConfigDir,
		"codefiles":   &global.CodeDir,
		"logs":        &global.LogDir,
	}

	// 遍历map，设置路径并创建不存在的目录
	for name, dir := range dirs {
		*dir = filepath.Join(global.CurrentDir, name)
		if _, err := os.Stat(*dir); os.IsNotExist(err) {
			os.Mkdir(*dir, os.ModePerm)
		}
	}

	// 初始化Logger
	logFile, err := utils.InitializeLogger()
	if err != nil {
		log.Fatalf("[FeasOJ] Failed to initialize logger: %v", err)
	}
	defer utils.CloseLogger(logFile)

	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	if utils.ConnectSql() == nil {
		return
	}
	log.Println("[FeasOJ] MySQL initialization complete")

	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.ConsulAddress
	log.Println("[FeasOJ] Connecting to Consul...")
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Println("[FeasOJ] Error connecting to Consul: ", err)
		return
	}

	// 构建沙盒镜像
	if judge.BuildImage() {
		log.Println("[FeasOJ] SandBox builds successfully")
	} else {
		log.Println("[FeasOJ] SandBox builds fail, please make sure Docker is running and up to date")
		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	router.LoadRouter(r)

	// 预热容器池
	judge.InitializeContainerPool(config.MaxSandbox)

	// 启动Judge任务处理协程
	go judge.ProcessJudgeTasks()

	startServer := func(protocol, address, certFile, keyFile string) {
		for {
			var err error
			if protocol == "http" {
				err = r.Run(address)
			} else {
				err = r.RunTLS(address, certFile, keyFile)
			}
			if err != nil {
				log.Printf("[FeasOJ] Server start error: %v\n", err)
				os.Exit(0)
			}
		}
	}

	if config.EnableHTTPS {
		go startServer("https", fmt.Sprintf("%s:%d", config.ServiceAddress, config.ServicePort), config.ServerCertPath, config.ServerKeyPath)
	} else {
		go startServer("http", fmt.Sprintf("%s:%d", config.ServiceAddress, config.ServicePort), "", "")
	}

	// 注册JudgeCore
	utils.RegService(consulClient)

	// 监听终端输入
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if scanner.Text() == "exit" || scanner.Text() == "EXIT" {
				log.Println("[FeasOJ] The server is being shut down, please be patient to wait for the container to be closed")
				os.Exit(0)
			}
		}
	}()

	// 等待中断信号关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("[FeasOJ] Input 'exit' or Ctrl+C to stop the server")
	<-quit

	log.Println("[FeasOJ] The server is shutting down, please be patient to wait for the container to be closed")
	judge.ShutdownContainerPool()
	utils.CloseLogger(logFile)
}
