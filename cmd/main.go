/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-01 23:22:18
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 04:00:42
 */
package main

import (
	"flag"

	"github.com/ossm-org/orchid/api"
	. "github.com/ossm-org/orchid/internal"
	"github.com/ossm-org/orchid/pkg"
)

var (
	port  int    // frontend host port number
	dev   bool   // whether is development mod or not, default to false
	level string // logging level, default to error.
)

func main() {
	flag.IntVar(&port, "port", 8000, "Frontend host port")
	flag.BoolVar(&dev, "dev", false, "Set development mod, default to false")
	flag.StringVar(&level, "level", "error", "Logging level")
	flag.Parse()

	server := pkg.NewServer(port)
	api.Rout(server.Root)
	Logger.Fatal(server.Run())
}
