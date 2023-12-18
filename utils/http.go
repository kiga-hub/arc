package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ResponseStatus response status
type ResponseStatus string

const (
	//ResponseStatusOK coresponse status
	ResponseStatusOK = "OK"
	//ResponseStatusError response status
	ResponseStatusError = "Error"
)

// Response -
type Response struct {
	Status ResponseStatus `json:"status" swagger:"required"`
	ErrStr string         `json:"err,omitempty"`
	Data   interface{}    `json:"data,omitempty"`
}

// ResponseV2 HTTP response , v2 version
type ResponseV2 struct {
	Code int         `json:"code" swagger:"required"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// ResponseWithoutData response without data
type ResponseWithoutData struct {
	Status ResponseStatus `json:"status" swagger:"required"`
	ErrStr string         `json:"err,omitempty"`
}

// GetJSONResponse get json response
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
