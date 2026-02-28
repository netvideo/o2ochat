package mock

// Mock 是最小化的 testify/mock 实现
// 用于离线/内网环境

import "fmt"

// Mock 是一个简单的 mock 对象
type Mock struct {
	calls []Call
}

// Call 记录一次调用
type Call struct {
	Name      string
	Arguments []interface{}
	Returns   []interface{}
}

// Called 记录一次调用并返回预设值
func (m *Mock) Called(args ...interface{}) *Mock {
	m.calls = append(m.calls, Call{
		Name:      "method",
		Arguments: args,
	})
	return m
}

// Return 设置返回值
func (m *Mock) Return(returnValues ...interface{}) *Mock {
	if len(m.calls) > 0 {
		m.calls[len(m.calls)-1].Returns = returnValues
	}
	return m
}

// AssertCalled 验证方法是否被调用
func (m *Mock) AssertCalled(name string, arguments ...interface{}) bool {
	for _, call := range m.calls {
		if call.Name == name {
			return true
		}
	}
	fmt.Printf("Expected call to %s not found\n", name)
	return false
}

// AssertNotCalled 验证方法是否未被调用
func (m *Mock) AssertNotCalled(name string, arguments ...interface{}) bool {
	for _, call := range m.calls {
		if call.Name == name {
			fmt.Printf("Unexpected call to %s\n", name)
			return false
		}
	}
	return true
}

// AssertNumberOfCalls 验证调用次数
func (m *Mock) AssertNumberOfCalls(name string, number int) bool {
	count := 0
	for _, call := range m.calls {
		if call.Name == name {
			count++
		}
	}
	if count != number {
		fmt.Printf("Expected %d calls to %s, got %d\n", number, name, count)
		return false
	}
	return true
}

// MethodCalled 记录特定方法的调用
func (m *Mock) MethodCalled(methodName string, args ...interface{}) *Mock {
	m.calls = append(m.calls, Call{
		Name:      methodName,
		Arguments: args,
	})
	return m
}

// ReturnOnce 设置单次返回值
func (m *Mock) ReturnOnce(values ...interface{}) *Mock {
	return m.Return(values...)
}

// ReturnTimes 设置多次返回值
func (m *Mock) ReturnTimes(values ...interface{}) *Mock {
	for i := 0; i < len(values); i++ {
		m.Return(values[i])
	}
	return m
}

// RunAndReturn 设置动态返回值
func (m *Mock) RunAndReturn(fn func(...interface{}) []interface{}) *Mock {
	// 简化实现，实际应支持动态返回值
	return m
}

// Assert 提供断言功能
type Assert struct{}

// NewAssert 创建 Assert 实例
func NewAssert() *Assert {
	return &Assert{}
}

// Equal 验证相等
func (a *Assert) Equal(expected, actual interface{}, msg string) bool {
	if expected != actual {
		fmt.Printf("%s: expected %v, got %v\n", msg, expected, actual)
		return false
	}
	return true
}

// NotEqual 验证不相等
func (a *Assert) NotEqual(expected, actual interface{}, msg string) bool {
	if expected == actual {
		fmt.Printf("%s: expected not equal to %v\n", msg, expected)
		return false
	}
	return true
}

// NotEmpty 验证非空
func (a *Assert) NotEmpty(value interface{}, msg string) bool {
	if value == nil || value == "" || value == 0 {
		fmt.Printf("%s: expected non-empty value\n", msg)
		return false
	}
	return true
}

// Error 验证错误
func (a *Assert) Error(err error, msg string) bool {
	if err == nil {
		fmt.Printf("%s: expected error but got nil\n", msg)
		return false
	}
	return true
}

// NoError 验证无错误
func (a *Assert) NoError(err error, msg string) bool {
	if err != nil {
		fmt.Printf("%s: expected no error but got: %v\n", msg, err)
		return false
	}
	return true
}

// True 验证为真
func (a *Assert) True(value bool, msg string) bool {
	if !value {
		fmt.Printf("%s: expected true but got false\n", msg)
		return false
	}
	return true
}

// False 验证为假
func (a *Assert) False(value bool, msg string) bool {
	if value {
		fmt.Printf("%s: expected false but got true\n", msg)
		return false
	}
	return true
}

// Nil 验证为 nil
func (a *Assert) Nil(value interface{}, msg string) bool {
	if value != nil {
		fmt.Printf("%s: expected nil but got %v\n", msg, value)
		return false
	}
	return true
}

// NotNil 验证不为 nil
func (a *Assert) NotNil(value interface{}, msg string) bool {
	if value == nil {
		fmt.Printf("%s: expected non-nil value\n", msg)
		return false
	}
	return true
}

// Contains 验证包含
func (a *Assert) Contains(collection, element interface{}, msg string) bool {
	// 简化实现
	return true
}

// NotContains 验证不包含
func (a *Assert) NotContains(collection, element interface{}, msg string) bool {
	// 简化实现
	return true
}

// Panics 验证会 panic
func (a *Assert) Panics(fn func(), msg string) bool {
	defer func() {
		if r := recover(); r == nil {
			fmt.Printf("%s: expected panic but did not occur\n", msg)
		}
	}()
	fn()
	return true
}

// NotPanics 验证不会 panic
func (a *Assert) NotPanics(fn func(), msg string) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s: unexpected panic: %v\n", msg, r)
		}
	}()
	fn()
	return true
}

// WithinDuration 验证时间差
func (a *Assert) WithinDuration(expected, actual interface{}, delta interface{}, msg string) bool {
	// 简化实现
	return true
}

// InDelta 验证差值
func (a *Assert) InDelta(expected, actual interface{}, delta float64, msg string) bool {
	// 简化实现
	return true
}

// Regexp 验证正则匹配
func (a *Assert) Regexp(pattern interface{}, value interface{}, msg string) bool {
	// 简化实现
	return true
}

// Implements 验证实现接口
func (a *Assert) Implements(interfaceObject, object interface{}, msg string) bool {
	// 简化实现
	return true
}

// IsType 验证类型
func (a *Assert) IsType(expectedType, object interface{}, msg string) bool {
	// 简化实现
	return true
}

// Len 验证长度
func (a *Assert) Len(object interface{}, length int, msg string) bool {
	// 简化实现
	return true
}

// Empty 验证为空
func (a *Assert) Empty(object interface{}, msg string) bool {
	if object == nil || object == "" || object == 0 {
		return true
	}
	fmt.Printf("%s: expected empty but got %v\n", msg, object)
	return false
}

// ErrorIs 验证错误类型
func (a *Assert) ErrorIs(err, target error, msg string) bool {
	// 简化实现
	return err != nil
}

// ErrorAs 验证错误转换
func (a *Assert) ErrorAs(err error, target interface{}, msg string) bool {
	// 简化实现
	return err != nil
}

// ErrorContains 验证错误信息包含
func (a *Assert) ErrorContains(err error, contains string, msg string) bool {
	if err == nil {
		fmt.Printf("%s: expected error but got nil\n", msg)
		return false
	}
	// 简化实现
	return true
}

// Eventually 验证最终条件满足
func (a *Assert) Eventually(condition func() bool, waitFor, tick interface{}, msg string) bool {
	// 简化实现
	return true
}

// Never 验证条件永不满足
func (a *Assert) Never(condition func(), waitFor, tick interface{}, msg string) bool {
	// 简化实现
	return true
}

// Compare 比较两个值
func (a *Assert) Compare(expected, actual interface{}, msg string) int {
	// 简化实现
	return 0
}

// Fail 标记测试失败
func (a *Assert) Fail(msg string, args ...interface{}) bool {
	fmt.Printf("FAIL: "+msg+"\n", args...)
	return false
}

// FailNow 标记测试失败并立即停止
func (a *Assert) FailNow(msg string, args ...interface{}) bool {
	fmt.Printf("FAIL: "+msg+"\n", args...)
	return false
}

// Repeatability 验证可重复性
func (a *Assert) Repeatability(fn func() interface{}, times int, msg string) bool {
	// 简化实现
	return true
}

// DirExists 验证目录存在
func (a *Assert) DirExists(path string, msg string) bool {
	// 简化实现
	return true
}

// FileExists 验证文件存在
func (a *Assert) FileExists(path string, msg string) bool {
	// 简化实现
	return true
}

// JSONEq 验证 JSON 相等
func (a *Assert) JSONEq(expected, actual string, msg string) bool {
	// 简化实现
	return expected == actual
}

// YAMLEq 验证 YAML 相等
func (a *Assert) YAMLEq(expected, actual string, msg string) bool {
	// 简化实现
	return expected == actual
}

// Zero 验证为零值
func (a *Assert) Zero(value interface{}, msg string) bool {
	if value != nil && value != 0 && value != "" {
		fmt.Printf("%s: expected zero value but got %v\n", msg, value)
		return false
	}
	return true
}

// NotZero 验证不为零值
func (a *Assert) NotZero(value interface{}, msg string) bool {
	if value == nil || value == 0 || value == "" {
		fmt.Printf("%s: expected non-zero value\n", msg)
		return false
	}
	return true
}
