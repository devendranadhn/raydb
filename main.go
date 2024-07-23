package main

import (
	"flag"
	"log"

	"ray/config"
	"ray/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "Host", "0.0.0.0", "host for the ray server")
	flag.IntVar(&config.Port, "Port", 7379, "port for the ray server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("Lighting up the Ray!!")
	server.RunASyncTCPServer()
}
