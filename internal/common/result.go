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

// ServiceResult 服务层统一结果
type ServiceResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// NewServiceResult 创建服务结果，与 chat-api 的 NewStudioServiceResult 实现方式一致
func NewServiceResult() *ServiceResult {
	return &ServiceResult{}
}

// SetCode 设置状态码
func (r *ServiceResult) SetCode(code int) {
	r.Code = code
}

// GetCode 获取状态码
func (r *ServiceResult) GetCode() int {
	return r.Code
}

// SetMessage 设置消息
func (r *ServiceResult) SetMessage(msg string) {
	r.Msg = msg
}

// GetMessage 获取消息
func (r *ServiceResult) GetMessage() string {
	return r.Msg
}

// GetData 获取数据
func (r *ServiceResult) GetData() any {
	return r.Data
}

// SetError 设置错误
func (r *ServiceResult) SetError(err *ServiceError, internalErr ...error) {
	r.Code = err.Code
	if len(internalErr) == 0 {
		r.Msg = err.Error()
	} else {
		r.Msg = fmt.Sprintf("%s, reason: %v", err.Error(), internalErr)
	}
}
