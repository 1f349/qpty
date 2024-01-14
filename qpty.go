package qpty

import (
	"context"
	"encoding/hex"
	"github.com/creack/pty"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
)

type Qpty struct {
	dock      *docker.Client
	container string
}

func New(client *docker.Client, container string) *Qpty {
	return &Qpty{
		dock:      client,
		container: container,
	}
}

func (q *Qpty) Exec(size *pty.Winsize) (*Exec, error) {
	iPty, _, err := pty.Open()
	if err != nil {
		return nil, err
	}
	e := &Exec{
		q:   q,
		pty: iPty,
	}
	if err := e.SetSize(size); err != nil {
		return nil, err
	}

	e.ir, e.iw = io.Pipe()
	e.or, e.ow = io.Pipe()

	return e, nil
}

type Exec struct {
	q      *Qpty
	pty    *os.File
	ir, or *io.PipeReader
	iw, ow *io.PipeWriter
}

func (e *Exec) Pty() *os.File {
	return e.pty
}

func (e *Exec) SetSize(winsize *pty.Winsize) error {
	return pty.Setsize(e.pty, winsize)
}

func (e *Exec) Send(p []byte) (int, error) {
	return e.iw.Write(p)
}

func (e *Exec) Run(shell string) error {
	execInst, err := e.q.dock.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{shell},
		Container:    e.q.container,
		User:         "root",
		WorkingDir:   "/",
		Context:      context.Background(),
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          true,
	})
	if err != nil {
		return err
	}

	go func() {
		_, _ = io.Copy(e.pty, e.or)
	}()

	go func() {
		r, w := io.Pipe()
		go io.Copy(hex.NewEncoder(w), e.pty)
		io.Copy(os.Stdout, r)
	}()

	return e.q.dock.StartExec(execInst.ID, docker.StartExecOptions{
		InputStream:  e.ir,
		OutputStream: e.ow,
		ErrorStream:  nil,
		Tty:          true,
		RawTerminal:  true,
		Context:      context.Background(),
	})
}
