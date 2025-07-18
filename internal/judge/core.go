package judge

import (
	"JudgeCore/internal/config"
	"JudgeCore/internal/global"
	"JudgeCore/internal/utils/sql"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/moby/go-archive"
)

// BuildImage 构建Sandbox
func BuildImage() bool {
	// 创建一个上下文
	ctx := context.Background()

	// 创建一个新的Docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("[FeasOJ] Error creating Docker client: ", err)
		return false
	}

	// 将Dockerfile目录打包成tar格式
	tar, err := archive.TarWithOptions(global.CurrentDir, &archive.TarOptions{})
	if err != nil {
		log.Println("[FeasOJ] Error creating tar: ", err)
		return false
	}

	// 设置镜像构建选项
	buildOptions := types.ImageBuildOptions{
		Context:    tar,                          // 构建上下文
		Dockerfile: "Sandbox",                    // Dockerfile文件名
		Tags:       []string{"judgecore:latest"}, // 镜像标签
	}

	log.Println("[FeasOJ] SandBox is being built...")
	// 构建Docker镜像
	buildResponse, err := cli.ImageBuild(ctx, tar, buildOptions)
	if err != nil {
		log.Println("[FeasOJ] Error building Docker image: ", err)
		return false
	}
	defer buildResponse.Body.Close()

	// 打印构建响应
	_, err = io.Copy(log.Writer(), buildResponse.Body)
	if err != nil {
		log.Printf("[FeasOJ] Error copying build response: %v", err)
	}

	return true
}

// StartContainer 启动Docker容器
func StartContainer() (string, error) {
	ctx := context.Background()

	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	// 配置容器配置
	containerConfig := &container.Config{
		Image: "judgecore:latest",
		Cmd:   []string{"sh"},
		Tty:   true,
	}

	// 配置主机配置
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:    config.GetSandboxMemory(),
			NanoCPUs:  int64(config.GetSandboxNanoCPUs() * 1e9),
			CPUShares: config.GetSandboxCPUShares(),
		},
		Binds: []string{
			global.CodeDir + ":/workspace", // 挂载文件夹
		},
		AutoRemove: true, // 容器退出后自动删除
		CapDrop:    []string{"ALL"},
	}

	// 创建容器
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	// 启动容器
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// ResetContainer 只清理任务专属的目录，而不影响其他任务
func ResetContainer(containerID, taskDir string) error {
	resetCmd := exec.Command("docker", "exec", containerID, "sh", "-c", fmt.Sprintf("rm -rf %s", taskDir))
	if err := resetCmd.Run(); err != nil {
		log.Printf("[FeasOJ] Error cleaning task directory %s in container %s: %v", taskDir, containerID, err)
		return err
	}
	return nil
}

// CompileAndRun 编译并运行代码
func CompileAndRun(filename string, containerID string) string {
	// 生成唯一任务目录（使用当前时间戳纳秒值）
	taskDir := fmt.Sprintf("/workspace/task_%d", time.Now().UnixNano())

	// 在容器内创建任务目录
	mkdirCmd := exec.Command("docker", "exec", containerID, "mkdir", "-p", taskDir)
	if err := mkdirCmd.Run(); err != nil {
		return "Internal Error"
	}

	// 将代码文件从挂载的workspace目录复制到任务目录中
	copyCmd := exec.Command("docker", "exec", containerID, "cp", fmt.Sprintf("/workspace/%s", filename), taskDir)
	if err := copyCmd.Run(); err != nil {
		return "Internal Error"
	}

	// 确保任务结束后清理任务目录
	defer func() {
		if err := ResetContainer(containerID, taskDir); err != nil {
			log.Printf("[FeasOJ] Reset task dir %s error: %v", taskDir, err)
		}
	}()

	ext := filepath.Ext(filename)
	var compileCmd *exec.Cmd

	// 解析题目ID
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename)) // 先去除扩展名
	parts := strings.Split(baseName, "_")
	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		return "Internal Error"
	}

	// 查询题目信息
	problem := sql.SelectProblemByPid(pid)

	// 解析时间限制和内存限制
	timeLimitStr := problem.Timelimit
	re := regexp.MustCompile(`\d+`)
	timeMatches := re.FindAllString(timeLimitStr, -1)
	if len(timeMatches) == 0 {
		return "Internal Error"
	}
	timeLimitSeconds, err := strconv.Atoi(timeMatches[0])
	if err != nil {
		return "Internal Error"
	}

	memoryLimitStr := problem.Memorylimit
	memMatches := re.FindAllString(memoryLimitStr, -1)
	if len(memMatches) == 0 {
		return "Internal Error"
	}
	memoryLimitMB, err := strconv.Atoi(memMatches[0])
	if err != nil {
		return "Internal Error"
	}
	memoryLimitKB := memoryLimitMB * 1024 // 转换为KB

	switch ext {
	case ".cpp":
		compileCmd = exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("g++ %s/%s -o %s/%s.out", taskDir, filename, taskDir, filename))
	case ".java":
		renameCmd := exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("mv %s/%s %s/Main.java", taskDir, filename, taskDir))
		if err := renameCmd.Run(); err != nil {
			return "Compile Failed"
		}
		// 编译Java代码
		compileCmd = exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("javac %s/Main.java", taskDir))
	case ".rs":
		exeName := strings.TrimSuffix(filename, ".rs")
		compileCmd = exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("rustc %s/%s -o %s/%s", taskDir, filename, taskDir, exeName))
	case ".php":
		compileCmd = exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("php -l %s/%s", taskDir, filename))
	case ".pas":
		// FPC编译Pas
		exeName := strings.TrimSuffix(filename, ".pas")
		compileCmd = exec.Command("docker", "exec", containerID, "sh", "-c",
			fmt.Sprintf("fpc -v0 -O2 %s/%s -o%s/%s", taskDir, filename, taskDir, exeName))
	default:

	}

	if compileCmd != nil {
		if err := compileCmd.Run(); err != nil {
			return "Compile Failed"
		}
	}

	testCases := sql.SelectTestCasesByPid(pid)
	for _, testCase := range testCases {
		// 每个测试用例使用独立的context，超时时间为题目限制+1秒缓冲
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimitSeconds+1)*time.Second)
		defer cancel()

		var cmdStr string
		// TODO: Java的内存限制需要调整修复
		switch ext {
		case ".cpp":
			cmdStr = fmt.Sprintf("ulimit -v %d && timeout -s SIGKILL %ds %s/%s.out", memoryLimitKB, timeLimitSeconds, taskDir, filename)
		case ".java":
			cmdStr = fmt.Sprintf("timeout -s SIGKILL %ds java -cp %s Main", timeLimitSeconds, taskDir)
		case ".py":
			cmdStr = fmt.Sprintf("ulimit -v %d && timeout -s SIGKILL %ds python %s/%s", memoryLimitKB, timeLimitSeconds, taskDir, filename)
		case ".rs":
			exeName := strings.TrimSuffix(filename, ".rs")
			cmdStr = fmt.Sprintf("ulimit -v %d && timeout -s SIGKILL %ds %s/%s",
				memoryLimitKB, timeLimitSeconds, taskDir, exeName)
		case ".php":
			cmdStr = fmt.Sprintf("ulimit -v %d && timeout -s SIGKILL %ds php %s/%s",
				memoryLimitKB, timeLimitSeconds, taskDir, filename)
		case ".pas":
			exeName := strings.TrimSuffix(filename, ".pas")
			cmdStr = fmt.Sprintf("ulimit -v %d && timeout -s SIGKILL %ds %s/%s",
				memoryLimitKB, timeLimitSeconds, taskDir, exeName)
		default:
			return "Failed"
		}

		runCmd := exec.CommandContext(ctx, "docker", "exec", "-i", containerID, "sh", "-c", cmdStr)
		runCmd.Stdin = strings.NewReader(testCase.InputData)
		output, err := runCmd.CombinedOutput()

		if ctx.Err() == context.DeadlineExceeded {
			return "Time Limit Exceeded"
		}

		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode := exitErr.ExitCode()
				switch exitCode {
				case 124: // timeout触发
					return "Time Limit Exceeded"
				case 137: // SIGKILL，可能是内存超限
					return "Memory Limit Exceeded"
				default:

				}
			}
			return "Failed"
		}

		// 添加调试信息
		expectedOutput := strings.TrimSpace(testCase.OutputData)
		actualOutput := strings.TrimSpace(string(output))

		if actualOutput != expectedOutput {
			return "Wrong Answer"
		}
	}

	return "Accepted"
}

// TerminateContainer 终止并删除Docker容器
func TerminateContainer(containerID string) bool {
	ctx := context.Background()

	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// 终止容器
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		panic(err)
	}

	return true
}
