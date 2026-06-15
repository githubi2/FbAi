package models

// BaseResponse 标准响应格式，与前端 art-design-pro 的 BaseResponse 完全对齐
// 前端定义: src/types/common/response.ts -> { code: number, msg: string, data: T }
type BaseResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// PageData 分页数据
type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// PageRequest 分页请求参数
type PageRequest struct {
	Page int    `json:"page" form:"page"`
	Size int    `json:"size" form:"size"`
	Sort string `json:"sort" form:"sort"`
}

// 预定义响应码
const (
	CodeSuccess      = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// Success 成功响应
func Success(data interface{}) BaseResponse {
	return BaseResponse{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	}
}

// SuccessWithMsg 带消息的成功响应
func SuccessWithMsg(msg string, data interface{}) BaseResponse {
	return BaseResponse{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	}
}

// Error 错误响应
func Error(code int, msg string) BaseResponse {
	return BaseResponse{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(code int, msg string, data interface{}) BaseResponse {
	return BaseResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// PageSuccess 分页成功响应
func PageSuccess(list interface{}, total int64, page, size int) BaseResponse {
	return Success(PageData{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
