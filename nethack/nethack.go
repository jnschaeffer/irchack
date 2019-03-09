package nethack

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

type NetHack struct {
	socketPath string
	socket     net.Conn
	process    *exec.Cmd
	options    []string
	stdoutTee  io.Reader
	stderrTee  io.Reader
}

func NewNetHack(socketPath string, options ...string) *NetHack {
	return &NetHack{
		socketPath: socketPath,
		options:    options,
	}
}

func (n *NetHack) Start() error {
	optionsStr := fmt.Sprintf("NETHACKOPTIONS=%s", strings.Join(n.options, ","))
	procArgs := []string{
		"-N",
		n.socketPath,
		"nethack",
	}

	process := exec.Command("dtach", procArgs...)
	process.Env = append(os.Environ(), optionsStr)

	stdoutPipe, errStdout := process.StdoutPipe()
	if errStdout != nil {
		return errStdout
	}

	stderrPipe, errStderr := process.StderrPipe()
	if errStderr != nil {
		return errStderr
	}

	n.stdoutTee = io.TeeReader(stdoutPipe, os.Stdout)
	n.stderrTee = io.TeeReader(stderrPipe, os.Stderr)

	errStart := process.Start()
	if errStart != nil {
		return errStart
	}

	time.Sleep(time.Second)

	socket, errDial := net.Dial("unix", n.socketPath)
	if errDial != nil {
		return errDial
	}

	n.process = process
	n.socket = socket

	return nil
}

func (n *NetHack) Close() error {
	errSocket := n.socket.Close()
	errProcess := n.process.Process.Signal(os.Kill)

	switch {
	case errProcess != nil:
		return errProcess
	case errSocket != nil:
		return errSocket
	default:
		return nil
	}
}

func (n *NetHack) Wait() error {
	errWait := n.process.Wait()

	return errWait
}

func (n *NetHack) Write(b []byte) (int, error) {
	msgLen := len(b)
	if msgLen > 255 {
		return -1, fmt.Errorf("sorry, I don't know how to send messages that long")
	}
	msg := []byte{0, byte(msgLen)}
	msg = append(msg, b...)

	return n.socket.Write(msg)
}
