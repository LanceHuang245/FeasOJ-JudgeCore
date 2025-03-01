package global

import "encoding/xml"

// 测试样例请求体
type TestCaseRequest struct {
	InputData  string `json:"input"`
	OutputData string `json:"output"`
}

// 配置文件结构体
type Config struct {
	XMLName   xml.Name  `xml:"config"`
	SqlConfig SqlConfig `xml:"sqlConfig"`
}

// MySQL数据库连接信息
type SqlConfig struct {
	DbName     string `xml:"dbname"`
	DbUser     string `xml:"dbuser"`
	DbPassword string `xml:"dbpassword"`
	DbAddress  string `xml:"dbaddress"`
}

type JudgeResultMessage struct {
	UserID    int    `json:"user_id"`
	ProblemID int    `json:"problem_id"`
	Status    string `json:"status"`
}
