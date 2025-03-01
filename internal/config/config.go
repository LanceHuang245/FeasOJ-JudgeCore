package config

import (
	"encoding/xml"
	"fmt"
	"log"
	"main/internal/global"
	"os"
	"path/filepath"
)

// 写入配置到文件
func WriteConfigToFile(filePath string, config global.Config) error {
	configXml, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, configXml, 0644)
}

// 读取配置文件
func ReadConfigFromFile(filePath string) (global.Config, error) {
	var config global.Config
	configFile, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer configFile.Close()
	err = xml.NewDecoder(configFile).Decode(&config)
	return config, err
}

// 初始化配置文件
func InitConfig() {
	filePath := filepath.Join(global.ConfigDir, "config.xml")
	// 判断是否有config.xml文件，没有则输入
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		InputSqlInfo()
	}
}

// 初始化并保存Sql配置
func InputSqlInfo() bool {
	var sqlConfig global.SqlConfig
	log.Println("[FeasOJ] Please input the MySQL connection configuration：")
	fmt.Print("[FeasOJ] Database address with port: ")
	fmt.Scanln(&sqlConfig.DbAddress)
	fmt.Print("[FeasOJ] Database name: ")
	fmt.Scanln(&sqlConfig.DbName)
	fmt.Print("[FeasOJ] Database user: ")
	fmt.Scanln(&sqlConfig.DbUser)
	fmt.Print("[FeasOJ] Database password: ")
	fmt.Scanln(&sqlConfig.DbPassword)
	log.Println("[FeasOJ] Saving the connection configuration...")

	filePath := filepath.Join(global.ConfigDir, "config.xml")
	config, _ := ReadConfigFromFile(filePath)
	config.SqlConfig = sqlConfig
	if err := WriteConfigToFile(filePath, config); err != nil {
		log.Println("[FeasOJ] Error saving SQL config: ", err)
		return false
	}
	return true
}

// 加载MySql配置
func LoadSqlConfig() string {
	filePath := filepath.Join(global.ConfigDir, "config.xml")
	config, err := ReadConfigFromFile(filePath)
	if err != nil {
		log.Println("[FeasOJ] Error loading MySQL config: ", err)
		return ""
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		config.SqlConfig.DbUser, config.SqlConfig.DbPassword, config.SqlConfig.DbAddress, config.SqlConfig.DbName)
	return dsn
}
