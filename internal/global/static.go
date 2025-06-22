package global

import "sync"

// 全局变量 - 本地配置文件路径
var ParentDir string
var LogDir string
var CodeDir string
var CurrentDir string
var CertDir string

// 全局变量 - 容器ID
var ContainerIDs sync.Map

// 全局配置变量
var AppConfigInstance *AppConfig
