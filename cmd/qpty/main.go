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
	proc, err := qpty.New(client, containerId, &pty.Winsize{Rows: 14, Cols: 11})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		milliType(proc, []byte(cmd+"\n"), 500*time.Millisecond)
	}()

	err = proc.Run(shell)
	if err != nil {
		log.Fatal(err)
	}
}

func milliType(exec *qpty.Qpty, p []byte, gap time.Duration) {
	for _, i := range p {
		exec.Send([]byte{i})
		time.Sleep(gap)
	}
}
