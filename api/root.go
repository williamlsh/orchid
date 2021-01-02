/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 03:59:54
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 04:06:50
 */
package api

import "github.com/gin-gonic/gin"

func Rout(g *gin.Engine) {
	user(g)
}
