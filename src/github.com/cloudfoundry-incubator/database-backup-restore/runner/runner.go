package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Run(cmd string, params []string, env map[string]string) ([]byte, []byte, error) {
	outb := bytes.NewBuffer([]byte{})
	errb := bytes.NewBuffer([]byte{})

	c := buildCommand(cmd, params, env, outb, errb)

	err := c.Run()

	return outb.Bytes(), errb.Bytes(), err
}

func RunWithStdin(cmd string, params []string, env map[string]string, stdin io.Reader) ([]byte, []byte, error) {
	outb := bytes.NewBuffer([]byte{})
	errb := bytes.NewBuffer([]byte{})

	c := buildCommand(cmd, params, env, outb, errb)
	c.Stdin = stdin

	err := c.Run()

	return outb.Bytes(), errb.Bytes(), err
}

func buildCommand(cmd string, params []string, env map[string]string, outb *bytes.Buffer, errb *bytes.Buffer) *exec.Cmd {
	command := exec.Command(cmd, params...)

	for key, value := range env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, value))
	}

	command.Stdout = io.MultiWriter(outb, os.Stdout)
	command.Stderr = io.MultiWriter(errb, os.Stderr)

	return command
}
