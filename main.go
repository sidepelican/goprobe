package main

import (
	"os"
	"fmt"
	"log"
	"github.com/takama/daemon"
)

const (
	// name of the service
	name         = "goprobe"
	description  = "caputre probe requests"
	dependencies = "network.target"
)

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return fmt.Sprintf("Usage: %v [install | remove | start | stop | status]", os.Args[0]), nil
		}
	}

	for {
		mainLoop()
	}

	// never happen, but need to complete code
	return "", nil
}

func main() {
	srv, err := daemon.New(name, description, dependencies)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		log.Println(status, "\nError:", err)
		os.Exit(1)
	}
	fmt.Println(status)
}