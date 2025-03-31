package judge

import (
	"log"
	"main/internal/config"
	"main/internal/global"
	"main/internal/utils"
	"main/internal/utils/sql"
	"strconv"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Task struct {
	UID  int
	PID  int
	Name string
}

// ProcessJudgeTasks 函数用于处理判题任务
func ProcessJudgeTasks() {
	// 连接到 RabbitMQ
	conn, ch, err := utils.ConnectRabbitMQ()
	if err != nil {
		log.Println("[FeasOJ] RabbitMQ connect error: ", err)
		return
	}
	log.Println("[FeasOJ] RabbitMQ connected")
	defer conn.Close()
	defer ch.Close()

	// 创建一个任务通道
	taskChan := make(chan Task)
	// 创建一个等待组
	var wg sync.WaitGroup

	// 创建多个 worker 协程
	for range config.MaxSandbox {
		wg.Add(1)
		go worker(taskChan, ch, &wg)
	}

	// 获取队列中的任务
	msgs, err := ch.Consume(
		"judgeTask", // 队列名称
		"",          // 消费者标签
		true,        // 自动应答
		false,       // 是否排他
		false,       // 是否持久化
		false,       // 是否等待
		nil,         // 额外参数
	)
	if err != nil {
		log.Panic("[FeasOJ] Failed to start consuming: ", err)
	}

	// 无限循环处理任务
	for msg := range msgs {
		taskData := string(msg.Body)
		// 将任务分割成用户ID和题目ID
		parts := strings.Split(taskData, "_")
		uid := parts[0]
		pid := strings.Split(parts[1], ".")[0]
		// 将用户ID和题目ID转换为整数
		uidInt, err := strconv.Atoi(uid)
		if err != nil {
			log.Panic(err)
		}
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			log.Panic(err)
		}

		// 将任务发送到任务通道
		taskChan <- Task{UID: uidInt, PID: pidInt, Name: taskData}
	}

	// 等待所有 worker 完成
	wg.Wait()
}

// worker 使用容器池执行任务
func worker(taskChan chan Task, ch *amqp.Channel, wg *sync.WaitGroup) {
	// 使用 defer 关键字，在函数结束时调用 wg.Done()，表示任务完成
	defer wg.Done()
	// 从任务通道中获取任务
	for task := range taskChan {
		// 从容器池中获取一个空闲容器
		containerID := AcquireContainer()
		// 将容器ID存储到全局变量中
		global.ContainerIDs.Store(task.Name, containerID)
		// 执行编译与运行
		result := CompileAndRun(task.Name, containerID)
		// TODO: 读取代码保存至SQL中并删除源文件
		// 更新判题状态
		sql.ModifyJudgeStatus(task.UID, task.PID, result)

		// 发送结果到消息队列
		resultMsg := global.JudgeResultMessage{
			UserID:    task.UID,
			ProblemID: task.PID,
			Status:    result,
		}

		if err := utils.PublishJudgeResult(ch, resultMsg); err != nil {
			log.Printf("[FeasOJ] Failed to publish result: %v", err)
		}

		// 将容器归还到池中（内部会先重置环境）
		ReleaseContainer(containerID)
	}
}
