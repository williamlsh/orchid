/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 01:26:39
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 03:58:37
 */
package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	port int
	Root *gin.Engine
}

// NewServer creates a new frontend.Server
func NewServer(port int) *HttpServer {
	return &HttpServer{
		port: port,
		Root: gin.New(),
	}
}

// Run starts the frontend server
func (s *HttpServer) Run() error {
	return s.Root.Run(":" + fmt.Sprint(s.port))
}
