/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 04:02:21
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 04:42:59
 */
package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ossm-org/orchid/internal/model"
	"github.com/ossm-org/orchid/pkg"
)

func user(g *gin.Engine) {
	v1 := g.Group("/user")
	{
		/**
		 * @api {post} /user/register 注册
		 * @apiName 注册
		 * @apiVersion 0.1.0
		 * @apiDescription 注册
		 * @apiSuccessExample {json}
		 * 		HTTP/1.1 200 OK
		 * 		{
		 * 			"success": true,
		 * 			"value": [{
		 * 				"objectId": "bb0894ff26c949898693f7bf6978c61a",
		 * 			}, ...]
		 * 		}
		 *
		 */
		v1.POST("/register", func(c *gin.Context) {
			req := model.User{}
			c.BindJSON(&req)
			req.Register()
		})
		/**
		 * @api {post} /user/veriCode 注册获取验证码
		 * @apiName 获取验证码
		 * @apiVersion 0.1.0
		 * @apiDescription 获取验证码
		 */
		v1.GET("/veriCode", func(c *gin.Context) {
			email := c.Param("email")
			err := pkg.SendMail(
				"1163388086@qq.com",
				"hwyjaqjakqadhahb",
				"smtp.qq.com:25",
				email,
				"orchid",
				"登录验证码",
				"1111",
				"html",
			)
			fmt.Errorf("", err)
		})
	}
}
