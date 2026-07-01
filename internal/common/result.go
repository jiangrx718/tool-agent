package common

// ServiceResult 服务层统一结果
type ServiceResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// NewServiceResult 创建成功结果
func NewServiceResult(data any) ServiceResult {
	return ServiceResult{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

// NewServiceError 创建错误结果
func NewServiceError(code int, msg string) ServiceResult {
	return ServiceResult{
		Code: code,
		Msg:  msg,
	}
}
