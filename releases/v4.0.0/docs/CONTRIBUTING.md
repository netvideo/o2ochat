# 贡献指南

欢迎参与 O2OChat 项目的开发！

## 开发环境设置

### 1. 安装工具

```bash
# Go 1.18+
go version

# Git
git --version

# 代码格式化
go install golang.org/x/tools/cmd/goimports@latest
```

### 2. 克隆项目

```bash
git clone https://github.com/netvideo/o2ochat.git
cd o2ochat
```

### 3. 安装依赖

```bash
go mod download
```

## 开发流程

### 1. 创建分支

```bash
# 基于 develop 分支
git checkout develop
git pull

# 创建特性分支
git checkout -b feature/your-feature-name
```

### 2. 编写代码

遵循以下规范：

- 使用有意义的变量和函数名
- 所有导出元素必须有文档注释
- 错误必须显式处理
- 遵循 Go 标准格式

### 3. 编写测试

```bash
# 为新增功能编写测试
# 确保测试覆盖率 > 80%

# 运行测试
go test ./... -v

# 运行特定模块测试
go test ./identity/... -v
```

### 4. 代码格式化

```bash
# 格式化代码
go fmt ./...

# 检查格式
gofmt -d .

# 运行 linter
golangci-lint run
```

### 5. 提交代码

```bash
# 添加更改
git add .

# 提交（使用约定式提交）
git commit -m "feat(module): description"

# 推送到远程
git push origin feature/your-feature-name
```

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### 类型说明

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具

### 示例

```
feat(identity): 添加 Ed25519 密钥生成
fix(transport): 修复 QUIC 连接内存泄漏
docs(signaling): 更新 API 文档
test(filetransfer): 添加分块测试
refactor(crypto): 重构加密接口
chore: 更新依赖版本
```

## 模块开发指南

### 1. 接口定义

每个模块必须有：

```
module/
├── interface.go    # 公共接口
├── types.go        # 数据结构
├── errors.go       # 错误类型
└── module.go       # 实现
```

### 2. 测试要求

```go
// 单元测试示例
func TestFunction(t *testing.T) {
    // 1. 准备测试数据
    // 2. 执行测试
    // 3. 验证结果
    // 4. 清理资源
}
```

### 3. 文档要求

```go
// FunctionName 函数功能说明
//
// 详细说明（可选）
//
// 参数:
//   - param1: 参数说明
//
// 返回:
//   - 返回值说明
//
// 示例:
//   result := FunctionName(param1)
func FunctionName(param1 Type) ReturnType {
    // 实现
}
```

## 代码审查清单

提交 PR 前请检查：

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 检查
- [ ] 所有测试通过
- [ ] 测试覆盖率达标
- [ ] 文档完整准确
- [ ] 无敏感信息泄露
- [ ] 遵循安全最佳实践

## 问题报告

### Bug 报告

请提供：

1. 问题描述
2. 复现步骤
3. 预期行为
4. 实际行为
5. 环境信息（OS、Go 版本等）
6. 日志/截图

### 功能请求

请提供：

1. 功能描述
2. 使用场景
3. 预期效果
4. 替代方案（如有）

## 测试指南

### 运行测试

```bash
# 所有测试
go test ./...

# 带覆盖率
go test ./... -cover

# 特定模块
go test ./identity/... -v

# 集成测试
go test ./tests/integration/... -v

# 性能测试
go test ./... -bench=. -benchmem
```

### 编写测试

```go
package module

import "testing"

func TestFeature(t *testing.T) {
    t.Run("subtest", func(t *testing.T) {
        // 测试代码
    })
}
```

## 发布流程

### 版本命名

遵循 [Semantic Versioning](https://semver.org/)：

- `MAJOR.MINOR.PATCH`
- 例如：`1.2.3`

### 发布检查清单

- [ ] 所有测试通过
- [ ] 文档更新完成
- [ ] 变更日志更新
- [ ] 版本号更新
- [ ] 构建验证通过

## 联系方式

- **Issues**: https://github.com/netvideo/o2ochat/issues
- **Discussions**: https://github.com/netvideo/o2ochat/discussions
- **Email**: dev@o2ochat.github.io

## 行为准则

- 尊重他人观点
- 建设性反馈
- 包容多样性
- 专业交流

---

感谢你的贡献！🎉
