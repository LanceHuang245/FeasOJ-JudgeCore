package global

// 测试样例请求体
type TestCaseRequest struct {
	InputData  string `json:"input"`
	OutputData string `json:"output"`
}

// JSON配置文件结构体
type AppConfig struct {
	// Consul配置
	Consul struct {
		Address     string `json:"address"`
		ServiceName string `json:"service_name"`
		ServiceID   string `json:"service_id"`
	} `json:"consul"`

	// RabbitMQ配置
	RabbitMQ struct {
		Address string `json:"address"`
	} `json:"rabbitmq"`

	// 服务器配置
	Server struct {
		Address     string `json:"address"`
		Port        int    `json:"port"`
		EnableHTTPS bool   `json:"enable_https"`
		CertPath    string `json:"cert_path"`
		KeyPath     string `json:"key_path"`
	} `json:"server"`

	// Docker沙盒配置
	Sandbox struct {
		Memory        int64   `json:"memory"`         // 内存限制 (字节)
		NanoCPUs      float64 `json:"nano_cpus"`      // CPU限制 (核心数)
		CPUShares     int64   `json:"cpu_shares"`     // CPU权重
		MaxConcurrent int     `json:"max_concurrent"` // 最大并发数
	} `json:"sandbox"`

	// MySQL配置
	Database struct {
		Address      string `json:"address"`
		Name         string `json:"name"`
		User         string `json:"user"`
		Password     string `json:"password"`
		MaxOpenConns int    `json:"max_open_conns"`
		MaxIdleConns int    `json:"max_idle_conns"`
		MaxLifeTime  int    `json:"max_life_time"`
	} `json:"database"`
}

// 判题结果信息结构体
type JudgeResultMessage struct {
	UserID    int    `json:"user_id"`
	ProblemID int    `json:"problem_id"`
	Status    string `json:"status"`
}

// 题目表: pid, difficulty, title, content, time_limit, memory_limit, input, output, contestid, is_visible
type Problem struct {
	Pid         int    `gorm:"comment:题目ID;primaryKey;autoIncrement"`
	Difficulty  string `gorm:"comment:难度;not null"`
	Title       string `gorm:"comment:题目标题;not null"`
	Content     string `gorm:"comment:题目详细;not null"`
	Timelimit   string `gorm:"comment:运行时间限制;not null"`
	Memorylimit string `gorm:"comment:内存大小限制;not null"`
	Input       string `gorm:"comment:输入样例;not null"`
	Output      string `gorm:"comment:输出样例;not null"`
	ContestID   int    `gorm:"comment:所属竞赛ID;not null"`
	IsVisible   bool   `gorm:"comment:是否可见;not null"`
}
