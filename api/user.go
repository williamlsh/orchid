/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 04:02:21
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 05:42:13
 */
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/ossm-org/orchid/internal/model"
	"github.com/ossm-org/orchid/pkg"
)

func user(g *gin.Engine) {
	v1 := g.Group("/user")
	{
		/**
		* @api {post} /user/register 注册
		* @apiGroup 用户部分
		* @apiName 注册
		* @apiVersion 0.1.0
		* @apiDescription 注册
		* @apiParam {string} account  帐号
		* @apiParam {string} email    邮箱
		* @apiParam {string} passwd   密码
		* @apiUse             FailResponse
		* @apiParamExample
		*      {
		* 		   "account":"test",
		* 		   "email":"test@qq.com",
		* 		   "passwd":"1123123"
		*      }
		* @apiSuccessExample {json}
		* 		HTTP/1.1 200 OK
		* 		{
		* 			"success": 1,
		* 			"data": {
		* 					     "account":"test",
		* 				         "email":"test@qq.com",
		* 						 "passwd":"1123123"
		* 				}
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
		* @apiGroup 用户部分
		* @apiName 注册获取验证码
		* @apiVersion 0.1.0
		* @apiDescription 注册获取验证码
		* @apiParam {string} email
		* @apiParamExample
		*			/user/veriCode?email=1163388086@qq.com
		* @apiUse             FailResponse
		* @apiUse             SuccessResponse
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
			c.JSON(200, model.ErrorIf(err))
		})
	}
}
