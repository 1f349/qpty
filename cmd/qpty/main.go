package main

import (
	"flag"
	"github.com/1f349/qpty"
	"github.com/creack/pty"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"time"
)

var dockerPath string
var containerId string
var shell string
var cmd string

func main() {
	flag.StringVar(&dockerPath, "docker", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVar(&containerId, "container", "", "container id")
	flag.StringVar(&shell, "shell", "/bin/sh", "shell path")
	flag.StringVar(&cmd, "cmd", "uptime", "command to execute")
	flag.Parse()

	client, err := docker.NewClient(dockerPath)
	if err != nil {
		log.Fatal(err)
	}
	machine := qpty.New(client, containerId)
	if err != nil {
		log.Fatal(err)
	}
	exec, err := machine.Exec(&pty.Winsize{Rows: 14, Cols: 11})
	if err != nil {
		return
	}

	go func() {
		time.Sleep(5 * time.Second)
		milliType(exec, []byte(cmd+"\n"), 500*time.Millisecond)
	}()

	err = exec.Run(shell)
	if err != nil {
		log.Fatal(err)
	}
}

func milliType(exec *qpty.Exec, p []byte, gap time.Duration) {
	for _, i := range p {
		exec.Send([]byte{i})
		time.Sleep(gap)
	}
}
