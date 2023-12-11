package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

//ResponseStatus 返回状态
type ResponseStatus string

const (
	//ResponseStatusOK 相应状态ok
	ResponseStatusOK = "OK"
	//ResponseStatusError 响应状态Error
	ResponseStatusError = "Error"
)

//Response 响应
type Response struct {
	Status ResponseStatus `json:"status" swagger:"required"`
	ErrStr string         `json:"err,omitempty"`
	Data   interface{}    `json:"data,omitempty"`
}

// ResponseV2 HTTP响应, v2版本
type ResponseV2 struct {
	Code int         `json:"code" swagger:"required"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

//ResponseWithoutData 响应没有数据
type ResponseWithoutData struct {
	Status ResponseStatus `json:"status" swagger:"required"`
	ErrStr string         `json:"err,omitempty"`
}

//GetJSONResponse 获取json响应
func GetJSONResponse(c echo.Context, err error, data interface{}) error {
	resp := Response{}
	if err != nil {
		resp.Status = ResponseStatusError
		resp.ErrStr = err.Error()
	} else {
		resp.Status = ResponseStatusOK
		resp.Data = data
	}
	return c.JSON(http.StatusOK, resp)
}
