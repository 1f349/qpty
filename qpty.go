package qpty

import (
	"context"
	"github.com/creack/pty"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
)

type Qpty struct {
	dock      *docker.Client
	container string
	pty, tty  *os.File
	ir, or    *io.PipeReader
	iw, ow    *io.PipeWriter
}

func New(client *docker.Client, container string, size *pty.Winsize) (*Qpty, error) {
	iPty, iTty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	q := &Qpty{
		dock:      client,
		container: container,
		pty:       iPty, tty: iTty,
	}
	if err := q.SetSize(size); err != nil {
		return nil, err
	}
	q.ir, q.iw = io.Pipe()
	q.or, q.ow = io.Pipe()
	return q, nil
}

func (q *Qpty) Pty() *os.File {
	return q.pty
}

func (q *Qpty) SetSize(winsize *pty.Winsize) error {
	return pty.Setsize(q.pty, winsize)
}

func (q *Qpty) Send(p []byte) (int, error) {
	return q.iw.Write(p)
}

func (q *Qpty) Run(shell string) error {
	execInst, err := q.dock.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{shell},
		Container:    q.container,
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
		_, _ = io.Copy(q.tty, q.or)
	}()

	return q.dock.StartExec(execInst.ID, docker.StartExecOptions{
		InputStream:  q.ir,
		OutputStream: q.ow,
		ErrorStream:  nil,
		Tty:          true,
		RawTerminal:  true,
		Context:      context.Background(),
	})
}
