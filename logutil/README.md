# logutil - 日志实用工具包

logutil 是一个为 kydenul/log 日志库提供实用工具函数的包。它提供了常见日志记录任务的便捷函数，帮助简化应用程序中的日志记录代码。

## 功能特性

### HTTP 请求日志工具

- `LogHTTPRequest()` - 记录 HTTP 请求详情
- `LogHTTPResponse()` - 记录 HTTP 响应详情，根据状态码自动选择日志级别

### 错误处理工具

- `LogError()` - 仅在错误不为 nil 时记录错误
- `FatalOnError()` - 在错误不为 nil 时记录致命错误并退出程序
- `CheckError()` - 检查并记录错误，返回是否有错误发生
- `Must()` - `FatalOnError()` 的简短别名

### 性能计时工具

- `Timer()` - 返回一个函数，调用时记录从创建到调用的耗时
- `TimeFunction()` - 执行函数并记录其执行时间

### 条件日志工具

- `InfoIf()` - 仅在条件为真时记录信息日志
- `ErrorIf()` - 仅在条件为真时记录错误日志
- `DebugIf()` - 仅在条件为真时记录调试日志
- `WarnIf()` - 仅在条件为真时记录警告日志

### 上下文日志工具

- `WithRequestID()` - 从上下文中提取请求 ID 并创建自动包含请求 ID 的日志包装器

### 异常处理工具

- `LogPanic()` - 恢复 panic 并记录为错误，然后重新 panic
- `LogPanicAsError()` - 恢复 panic 并记录为错误，不重新 panic

### 应用生命周期工具

- `LogStartup()` - 记录应用启动信息
- `LogShutdown()` - 记录应用关闭信息

## 使用示例

### HTTP 请求日志

```go
import (
    "github.com/kydenul/log"
    "github.com/kydenul/log/logutil"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
    logger := log.NewLog(nil)
    start := time.Now()
    
    // 记录请求
    logutil.LogHTTPRequest(logger, r)
    
    // 处理请求...
    statusCode := 200
    
    // 记录响应
    duration := time.Since(start)
    logutil.LogHTTPResponse(logger, r, statusCode, duration)
}
```

### 错误处理

```go
func processData() error {
    logger := log.NewLog(nil)
    
    data, err := loadData()
    if logutil.CheckError(logger, err, "Failed to load data") {
        return err
    }
    
    // 继续处理...
    return nil
}

func criticalOperation() {
    logger := log.NewLog(nil)
    
    err := performCriticalTask()
    logutil.FatalOnError(logger, err, "Critical task failed")
    
    // 如果没有错误，继续执行...
}
```

### 性能计时

```go
func databaseQuery() {
    logger := log.NewLog(nil)
    
    // 方法1：使用 Timer
    defer logutil.Timer(logger, "database_query")()
    
    // 执行数据库查询...
}

func processData() {
    logger := log.NewLog(nil)
    
    // 方法2：使用 TimeFunction
    logutil.TimeFunction(logger, "data_processing", func() {
        // 数据处理逻辑...
    })
}
```

### 条件日志

```go
func validateInput(input string) error {
    logger := log.NewLog(nil)
    debugMode := os.Getenv("DEBUG") == "true"
    
    logutil.InfoIf(logger, debugMode, "Validating input", "input_length", len(input))
    
    if input == "" {
        logutil.ErrorIf(logger, true, "Input validation failed", "reason", "empty_input")
        return errors.New("input cannot be empty")
    }
    
    return nil
}
```

### 请求 ID 日志

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    logger := log.NewLog(nil)
    
    // 从上下文中获取请求 ID 并创建包装器
    ctx := context.WithValue(r.Context(), "request_id", generateRequestID())
    requestLogger := logutil.WithRequestID(logger, ctx)
    
    // 所有日志调用都会自动包含请求 ID
    requestLogger.Info("Processing request")
    requestLogger.Error("Request failed")
}
```

### 异常处理

```go
func riskyOperation() {
    logger := log.NewLog(nil)
    
    // 恢复 panic 并记录，但不重新 panic
    defer logutil.LogPanicAsError(logger, "risky_operation")
    
    // 可能会 panic 的代码...
    panic("something went wrong")
    
    // 这行代码不会执行，但程序不会崩溃
}
```

### 应用生命周期

```go
func main() {
    logger := log.NewLog(nil)
    startTime := time.Now()
    
    // 记录启动
    logutil.LogStartup(logger, "my-service", "v1.0.0", 8080)
    
    // 应用逻辑...
    
    // 记录关闭
    uptime := time.Since(startTime)
    logutil.LogShutdown(logger, "my-service", uptime)
}
```

## 设计原则

1. **空安全**: 所有函数都能安全处理 nil logger 和 nil 参数
2. **非侵入性**: 不修改原始日志库的行为，只是提供便捷包装
3. **性能友好**: 避免不必要的操作，如条件检查在日志记录前进行
4. **一致性**: 遵循原日志库的命名约定和参数模式
5. **实用性**: 专注于解决实际开发中的常见日志记录需求

## 兼容性

- 兼容任何实现 `log.Logger` 接口的日志器
- 不依赖特定的日志实现细节
- 可以与现有代码无缝集成

## 测试

运行测试：

```bash
go test ./logutil -v
```

查看测试覆盖率：

```bash
go test ./logutil -cover
```
