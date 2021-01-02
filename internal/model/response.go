/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 05:05:48
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 05:37:25
 */
package model

type Response struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
}

func Fail(err error) Response {
	return Response{
		Code: 0,
		Msg:  err.Error(),
	}
}
func ErrorIf(err error) Response {
	if nil != err {
		return Response{
			Code: 0,
			Msg:  err.Error(),
		}
	} else {
		return Response{
			Code: 1,
		}
	}
}
func Succ(data interface{}) Response {
	return Response{
		Code: 1,
		Data: data,
	}
}
