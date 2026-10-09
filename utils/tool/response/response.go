package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess      = 0
	CodeParameterErr = 400
	CodeInternalErr  = 500
)

// 统一的默认提示
const (
	MsgSuccess = "success"
	MsgFailed  = "failed"
)

// withDefault 当 msg 为空时返回默认失败提示，否则原样返回
func withDefault(msg, def string) string {
	if msg == "" {
		return def
	}
	return msg
}

// Response 统一响应结构
type Response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Data   any    `json:"data,omitempty"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
	Count  *int   `json:"count,omitempty"`
}

// Successful 返回成功响应
func Successful(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: MsgSuccess, Data: data})
}

// SuccessfulWithPagination 返回带分页的成功响应
func SuccessfulWithPagination(ctx *gin.Context, data any, offset *int, limit *int, count *int) {
	ctx.JSON(http.StatusOK, Response{
		Code:   CodeSuccess,
		Msg:    MsgSuccess,
		Data:   data,
		Count:  count,
		Limit:  limit,
		Offset: offset,
	})
}

// ParameterError 参数错误
// err 不为空时使用其错误信息，否则默认 failed
func ParameterError(ctx *gin.Context, err error) {
	msg := MsgFailed
	if err != nil {
		msg = err.Error()
	}
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeParameterErr, Msg: msg})
}

// InternalError 内部错误
// 可通过 msgs 传入错误信息；为空时默认 failed
func InternalError(ctx *gin.Context, msgs ...string) {
	msg := MsgFailed
	if len(msgs) > 0 && msgs[0] != "" {
		msg = msgs[0]
	}
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: CodeInternalErr, Msg: msg})
}

// Failed 返回失败响应
// message 为空时默认 failed
func Failed(ctx *gin.Context, code int, message string, data any) {
	ctx.AbortWithStatusJSON(http.StatusOK, Response{Code: code, Msg: withDefault(message, MsgFailed), Data: data})
}
