package main

import (
	"errors"

	"github.com/kydenul/log"
)

func main() {
	// 创建一个自定义配置的 logger
	logger := log.NewBuilder().
		Level("debug").
		Format("console").
		Directory("./logs").
		Prefix("[DualMode] ").
		ConsoleOutput(true).
		Build()

	// 演示双重调用模式
	demonstrateDualCalling(logger)

	// 演示不同创建方式都支持双重调用
	demonstrateAllCreationMethods()

	// 同步日志
	logger.Sync()
}

func demonstrateDualCalling(logger *log.Log) {
	log.Info("=== 双重调用模式演示 ===")

	// 方式1: 实例方法调用
	logger.Info("这是实例方法调用")
	logger.Warn("实例方法警告", "component", "auth")
	logger.Errorf("实例方法错误: %v", errors.New("示例错误"))

	// 方式2: 全局函数调用（使用相同配置）
	log.Info("这是全局函数调用")
	log.Warn("全局函数警告", "component", "database")
	log.Errorf("全局函数错误: %v", errors.New("示例错误"))

	// 结构化日志
	logger.Infow("实例方法结构化日志", "user_id", 123, "action", "login", "success", true)
	log.Infow("全局函数结构化日志", "user_id", 456, "action", "logout", "success", true)

	log.Info("注意：两种调用方式产生相同格式的日志输出")
}

func demonstrateAllCreationMethods() {
	log.Info("\n=== 所有创建方式都支持双重调用 ===")

	// 1. Quick 方式
	logger1 := log.Quick()
	logger1.Info("Quick方式 - 实例调用")
	log.Info("Quick方式 - 全局调用")

	// 2. 预设方式
	logger2 := log.WithPreset(log.DevelopmentPreset())
	logger2.Info("开发预设 - 实例调用")
	log.Info("开发预设 - 全局调用")

	// 3. NewLog 方式
	opts := log.NewOptions()
	opts.Level = "info"
	opts.Prefix = "[Custom] "
	logger3 := log.NewLog(opts)
	logger3.Info("NewLog方式 - 实例调用")
	log.Info("NewLog方式 - 全局调用")

	log.Info("最后创建的 logger (logger3) 现在是全局默认 logger")
}
