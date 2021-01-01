package main

import (
	"flag"

	"github.com/ossm-org/orchid/pkg/logging"
	"github.com/ossm-org/orchid/services/frontend"
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

	logger := logging.NewLogger(level, dev)
	server := frontend.NewServer(logger)
	logger.Fatal(server.Run())
}
