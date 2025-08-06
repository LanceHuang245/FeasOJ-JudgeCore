package judge

import (
	"JudgeCore/internal/config"
	"JudgeCore/internal/global"
	"JudgeCore/internal/utils"
	"JudgeCore/internal/utils/sql"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Task struct {
	UID  int
	PID  int
	Name string
}

// ProcessJudgeTasks 函数用于处理判题任务
func ProcessJudgeTasks() {
	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	// 若断开则自动重连
	for {
		conn, ch, err = utils.ConnectRabbitMQ()
		if err != nil {
			log.Println("[FeasOJ] RabbitMQ connect error, retrying in 3s: ", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("[FeasOJ] RabbitMQ connected")
		break
	}
	defer conn.Close()
	defer ch.Close()

	// 创建一个任务通道
	taskChan := make(chan Task)
	// 创建一个等待组
	var wg sync.WaitGroup

	// 创建多个 worker 协程
	for range config.GetMaxSandbox() {
		wg.Add(1)
		go worker(taskChan, ch, &wg)
	}

	for {
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
			log.Println("[FeasOJ] Failed to start consuming, retrying in 3s: ", err)
			time.Sleep(3 * time.Second)
			// 重新连接
			conn.Close()
			ch.Close()
			for {
				conn, ch, err = utils.ConnectRabbitMQ()
				if err != nil {
					log.Println("[FeasOJ] RabbitMQ reconnect error, retrying in 3s: ", err)
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}
			continue
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
		break
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
