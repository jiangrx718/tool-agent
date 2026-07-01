package common

import "fmt"

// ServiceError 服务层错误
type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// ServiceResult 服务层统一结果接口，对齐 chat-api 的 common.ServiceResult
type ServiceResult interface {
	SetCode(code int)
	GetCode() int
	SetMessage(msg string)
	GetMessage() string
	GetData() any
	SetError(err *ServiceError, internalErr ...error)
}

// BaseServiceResult 服务层结果基类，对齐 chat-api 的 common.BaseServiceResult
type BaseServiceResult struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data,omitempty"`
}

// NewServiceResult 创建服务结果，返回基类指针
func NewServiceResult() *BaseServiceResult {
	return &BaseServiceResult{}
}

// SetCode 设置状态码
func (r *BaseServiceResult) SetCode(code int) {
	r.Code = code
}

// GetCode 获取状态码
func (r *BaseServiceResult) GetCode() int {
	return r.Code
}

// SetMessage 设置消息
func (r *BaseServiceResult) SetMessage(msg string) {
	r.Message = msg
}

// GetMessage 获取消息
func (r *BaseServiceResult) GetMessage() string {
	return r.Message
}

// GetData 获取数据
func (r *BaseServiceResult) GetData() any {
	return r.Data
}

// SetError 设置错误
func (r *BaseServiceResult) SetError(err *ServiceError, internalErr ...error) {
	r.Code = err.Code
	if len(internalErr) == 0 {
		r.Message = err.Error()
	} else {
		r.Message = fmt.Sprintf("%s, reason: %v", err.Error(), internalErr)
	}
}
