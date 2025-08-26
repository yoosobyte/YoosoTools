package entity

import (
	"encoding/json"
)

// R 统一响应结构体
type R struct {
	Code int         `json:"code"`           // 状态码：200成功，其他为错误
	Msg  string      `json:"msg"`            // 消息描述
	Data interface{} `json:"data,omitempty"` // 数据，omitempty 表示如果为空则不输出
}

// 成功状态码
const (
	CodeSuccess = 200
)

// 错误状态码
const (
	CodeError        = 500
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeNotFound     = 404
)

func Success() *R {
	return &R{
		Code: CodeSuccess,
		Msg:  "操作成功",
		Data: nil,
	}
}

func SuccessStr() string {
	r := &R{
		Code: CodeSuccess,
		Msg:  "操作成功",
		Data: nil,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

func SuccessOnlyData(data interface{}) *R {
	return &R{
		Code: CodeSuccess,
		Msg:  "操作成功",
		Data: data,
	}
}

func SuccessOnlyDataStr(data interface{}) string {
	r := &R{
		Code: CodeSuccess,
		Msg:  "操作成功",
		Data: data,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

func SuccessOnlyMsg(msg string) *R {
	return &R{
		Code: CodeSuccess,
		Msg:  msg,
		Data: nil,
	}
}

func SuccessOnlyMsgStr(msg string) string {
	r := &R{
		Code: CodeSuccess,
		Msg:  msg,
		Data: nil,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

func SuccessWithAll(msg string, data interface{}) *R {
	return &R{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	}
}
func SuccessWithAllStr(msg string, data interface{}) string {
	r := &R{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

// Error 错误响应
func Error(msg string) *R {
	return &R{
		Code: CodeError,
		Msg:  msg,
		Data: nil,
	}
}

func ErrorStr() string {
	r := &R{
		Code: CodeError,
		Msg:  "操作失败",
		Data: nil,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

func ErrorOnlyMsgStr(msg string) string {
	r := &R{
		Code: CodeError,
		Msg:  msg,
		Data: nil,
	}
	// 将对象转换为 JSON 字节切片
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

func ErrorOnlyDataStr(data interface{}) string {
	r := &R{
		Code: CodeError,
		Msg:  "操作成功",
		Data: data,
	}
	jsonBytes, _ := json.Marshal(r)
	return string(jsonBytes)
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(code int, msg string) *R {
	return &R{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

// BadRequest 请求参数错误
func BadRequest(msg string) *R {
	return ErrorWithCode(CodeBadRequest, msg)
}

// Unauthorized 未授权
func Unauthorized(msg string) *R {
	return ErrorWithCode(CodeUnauthorized, msg)
}

// NotFound 资源未找到
func NotFound(msg string) *R {
	return ErrorWithCode(CodeNotFound, msg)
}

// IsSuccess 判断响应是否成功
func (r *R) IsSuccess() bool {
	return r.Code == CodeSuccess
}
