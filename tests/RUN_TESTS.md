# 测试运行指南

## 环境要求

- Go 1.21+
- 正确配置 GOROOT 和 GOPATH

### GOROOT 配置

如果遇到 `go: cannot find GOROOT directory` 错误，请设置正确的 GOROOT：

```bash
# Linux - 设置正确的 GOROOT
export GOROOT=/usr/local/soft
export PATH=$GOROOT/bin:$PATH

# 验证 Go 安装
go version
# 输出: go version go1.25.4 linux/amd64
```

## 运行测试

### 1. 单元测试

```bash
# 运行所有单元测试
go test ./tests/unit/... -v

# 运行特定模块测试
go test ./tests/unit/identity/... -v
go test ./tests/unit/crypto/... -v
go test ./tests/unit/storage/... -v

# 快速测试（跳过耗时测试）
go test ./tests/unit/... -v -short

# 带覆盖率测试
go test ./tests/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 2. 集成测试

```bash
# 运行所有集成测试
go test ./tests/integration/... -v

# 运行特定集成测试
go test ./tests/integration/identity_signaling_test.go -v
go test ./tests/integration/crypto_storage_test.go -v
go test ./tests/integration/full_stack_test.go -v
```

### 3. 性能测试

```bash
# 运行所有性能基准测试
go test ./tests/performance/... -v -bench=.

# 运行特定基准测试
go test ./tests/performance/crypto_benchmark_test.go -v -bench=.
go test ./tests/performance/transport_benchmark_test.go -v -bench=.
go test ./tests/performance/filetransfer_benchmark_test.go -v -bench=.

# 生成性能报告
go test ./tests/performance/... -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
```

### 4. 端到端测试

```bash
# 运行端到端测试
go test ./tests/e2e/... -v

# 注意：端到端测试耗时较长，建议使用 -short 标志跳过
go test ./tests/e2e/... -v -short
```

### 5. 安全测试

```bash
# 运行安全测试
go test ./tests/security/... -v

# 运行特定安全测试
go test ./tests/security/crypto_security_test.go -v
go test ./tests/security/identity_security_test.go -v
```

### 6. 带 Race 检测

```bash
# 运行所有测试并检测竞态条件
go test ./... -race -v
```

## 测试分类和标签

| 标签 | 说明 | 命令 |
|------|------|------|
| unit | 单元测试 | `go test -run Unit` |
| integration | 集成测试 | `go test -tags=integration` |
| e2e | 端到端测试 | `go test -tags=e2e` |
| security | 安全测试 | `go test -tags=security` |
| performance | 性能测试 | `go test -bench=.` |
| short | 快速测试 | `go test -short` |

## 预期测试结果

### 单元测试
- 预期通过率：100%
- 覆盖率目标：≥80%

### 集成测试
- 预期通过率：100%
- 测试场景：模块间交互

### 性能测试
- 加密/解密：>100MB/s
- 签名/验证：>10,000 ops/s
- 文件分块：>500MB/s

### 安全测试
- 预期通过率：100%
- 无高危漏洞

## 常见问题

### GOROOT 错误
```bash
# 设置正确的 GOROOT
export GOROOT=$(go env GOROOT)
```

### 依赖缺失
```bash
# 安装依赖
go mod tidy
go mod download
```

### 测试超时
```bash
# 增加超时时间
go test -timeout 30m ./...
```

## 测试报告生成

```bash
# 生成 JUnit 格式报告
go test -v -json ./... > test-results.json

# 生成 HTML 报告
go test -v ./... | go2testium -o test-report.html

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 持续集成

### GitHub Actions 示例
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: '1.21'
    
    - name: Run unit tests
      run: go test ./tests/unit/... -v -race
    
    - name: Run integration tests
      run: go test ./tests/integration/... -v -tags=integration
    
    - name: Run security tests
      run: go test ./tests/security/... -v -tags=security
    
    - name: Generate coverage report
      run: |
        go test ./... -coverprofile=coverage.out
        go tool cover -func=coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v2
```

## 测试文件清单

### 单元测试 (50+ 文件)
- tests/unit/identity/identity_test.go
- tests/unit/integration_test.go
- [各模块内部测试文件...]

### 集成测试 (4 个文件)
- tests/integration/signaling_transport_test.go
- tests/integration/identity_signaling_test.go
- tests/integration/crypto_storage_test.go
- tests/integration/full_stack_test.go

### 性能测试 (3 个文件)
- tests/performance/crypto_benchmark_test.go
- tests/performance/transport_benchmark_test.go
- tests/performance/filetransfer_benchmark_test.go

### 端到端测试 (1 个文件)
- tests/e2e/full_flow_test.go

### 安全测试 (2 个文件)
- tests/security/crypto_security_test.go
- tests/security/identity_security_test.go

### Mock 对象 (4 个文件)
- tests/mocks/identity_mock.go
- tests/mocks/signaling_mock.go
- tests/mocks/transport_mock.go
- tests/mocks/filetransfer_mock.go

### 测试工具 (1 个文件)
- tests/utils/test_helpers.go

## 已知问题

### 模块编译问题

部分模块存在编译错误，需要先修复才能运行测试：

1. **media 模块** - 存在接口和结构体命名冲突
   - 问题：interface.go 中定义的接口与 codec.go, rtp_processor.go, session.go, video_processor.go 中的结构体同名
   - 状态：需要重构，将接口和实现分离

2. **signaling 模块** - 缺少依赖
   - 问题：缺少 github.com/gorilla/websocket
   - 解决：运行 `go get github.com/gorilla/websocket`

3. **filetransfer 模块** - 测试代码问题
   - 问题：测试引用未定义的函数 NewChunkManagerImpl, NewSchedulerImpl
   - 状态：需要修复测试代码

### 已验证通过的模块

- **identity 模块** ✅
  - 测试用例：28+
  - 覆盖率：86.5%
  - 状态：全部通过

## 联系支持

如有测试相关问题，请联系开发团队或提交 Issue。
