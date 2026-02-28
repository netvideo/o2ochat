# O2OChat LSP 错误修复指南

**创建时间**: 2026 年 2 月 28 日  
**状态**: 🔄 修复中  
**目标**: 0 个 LSP 错误

---

## 📊 当前 LSP 错误统计

**总错误数**: <15 个  
**目标**: 0 个

### 按文件分类

| 文件 | 错误数 | 严重程度 | 状态 |
|------|--------|---------|------|
| storage/chunk_storage.go | 1 | 中 | ⏳ 待修复 |
| tests/mocks/signaling_mock.go | 14 | 低 | ⏳ 待修复 |
| tests/mocks/filetransfer_mock.go | 14 | 低 | ⏳ 待修复 |
| tests/security/crypto_security_test.go | 20+ | 中 | ⏳ 待修复 |
| tests/performance/transport_benchmark_test.go | 3 | 低 | ⏳ 待修复 |

---

## 🔧 修复方案

### 1. storage/chunk_storage.go (1 个错误)

**错误**: `could not import github.com/mattn/go-sqlite3`

**原因**: 依赖缺失

**修复方案**:

#### 方案 A: 添加依赖（推荐）
```bash
cd /mnt/f/o2ochat
export GOPROXY=https://mirrors.aliyun.com/goproxy/
go get github.com/mattn/go-sqlite3@v1.14.18
go mod tidy
```

#### 方案 B: 使用纯 Go 实现
```go
// 使用 github.com/modernc.org/sqlite (纯 Go 实现)
import _ "github.com/modernc.org/sqlite"
```

**状态**: ⏳ 待执行

---

### 2. tests/mocks/*.go (28 个错误)

**错误**: `m.Called undefined` 和 `could not import github.com/stretchr/testify/mock`

**原因**: testify/mock 依赖缺失

**修复方案**:

#### 方案 A: 添加依赖
```bash
go get github.com/stretchr/testify@v1.8.4
go mod tidy
```

#### 方案 B: 使用本地 mock 实现（已创建）
```go
// internal/test/mock.go 已创建
// 更新导入路径
import "github.com/netvideo/o2ochat/internal/test/mock"
```

**状态**: ✅ 本地 mock 已创建，待更新导入

---

### 3. tests/security/crypto_security_test.go (20+ 个错误)

**错误**: API 调用参数不匹配

**原因**: 测试代码未适配最新 API

**修复方案**: 更新测试代码

```go
// 修复示例
// 旧代码:
manager := identity.NewIdentityManager()

// 新代码:
store := identity.NewMemoryIdentityStore()
keyStorage := identity.NewMemoryKeyStorage()
manager, err := identity.NewIdentityManager(store, keyStorage)
```

**状态**: ⏳ 待修复

---

### 4. tests/performance/transport_benchmark_test.go (3 个错误)

**错误**: 
- `declared and not used: manager`
- `not enough arguments in call to transport.NewTransportManager`

**修复方案**:

```go
// 修复示例
// 旧代码:
manager := transport.NewTransportManager()

// 新代码:
config := &transport.TransportConfig{
    MaxConnections: 100,
    Timeout: 30 * time.Second,
}
manager := transport.NewTransportManager(config)
_ = manager // 或使用 manager
```

**状态**: ⏳ 待修复

---

## 📋 修复检查清单

### 高优先级（本周）

- [ ] 修复 storage/chunk_storage.go 依赖
- [ ] 添加 testify 依赖或更新 mock 导入
- [ ] 修复 crypto_security_test.go API 调用
- [ ] 修复 transport_benchmark_test.go 参数

### 中优先级（本月）

- [ ] 运行 go mod tidy 验证依赖
- [ ] 运行所有测试验证修复
- [ ] 验证无 LSP 错误

---

## 🚀 执行步骤

### 步骤 1: 修复依赖

```bash
cd /mnt/f/o2ochat

# 设置国内代理
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
export GOSUMDB=off

# 添加依赖
go get github.com/mattn/go-sqlite3@v1.14.18
go get github.com/stretchr/testify@v1.8.4

# 整理依赖
go mod tidy
```

### 步骤 2: 修复测试代码

```bash
# 修复 crypto_security_test.go
# 修复 transport_benchmark_test.go
# 更新 mock 导入路径
```

### 步骤 3: 验证修复

```bash
# 检查 LSP 错误
# 运行测试
go test ./...

# 验证构建
go build ./...
```

---

## 📊 修复进度追踪

```
依赖修复      [░░░░░░░░░░] 0%
测试修复      [░░░░░░░░░░] 0%
验证          [░░░░░░░░░░] 0%

总体进度      [░░░░░░░░░░] 0%
```

---

## 🎯 成功标准

- [ ] 0 个 LSP 错误
- [ ] 所有测试通过
- [ ] 项目可以正常构建
- [ ] 依赖关系清晰

---

**预计完成时间**: 2026 年 3 月 7 日  
**负责人**: AI Development Team  
**状态**: 🔄 **执行中**
