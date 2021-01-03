/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 03:59:54
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 05:16:19
 */
package api

import "github.com/gin-gonic/gin"

/**
 * @apiDefine        FailResponse
 * @apiErrorExample  Response (fail):
 *     {
 *       "code": 0
 *       "Msg": "错误内容"
 *     }
 */

/**
 * @apiDefine        SuccessResponse
 * @apiSuccessExample  Response (success):
 *     {
 *       "code":"1"
 *       "data":"成功数据"
 *     }
 */

func Rout(g *gin.Engine) {
	user(g)
}
