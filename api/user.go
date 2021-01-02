/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 04:02:21
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 04:19:49
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
		v1.POST("/register", func(c *gin.Context) {
			req := model.User{}
			c.BindJSON(&req)
			req.Register()
		})
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
