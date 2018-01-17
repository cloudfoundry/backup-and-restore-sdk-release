package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Command struct {
	cmd    string
	params []string
	env    map[string]string
	stdin  io.Reader
}

func NewCommand(cmd string) Command {
	return Command{cmd: cmd}
}

func (c Command) WithParams(params ...string) Command {
	return Command{cmd: c.cmd, params: params, env: c.env, stdin: c.stdin}
}

func (c Command) WithEnv(env map[string]string) Command {
	return Command{cmd: c.cmd, params: c.params, env: env, stdin: c.stdin}
}

func (c Command) WithStdin(stdin io.Reader) Command {
	return Command{cmd: c.cmd, params: c.params, env: c.env, stdin: stdin}
}

func (c Command) Run() ([]byte, []byte, error) {
	outb := bytes.NewBuffer([]byte{})
	errb := bytes.NewBuffer([]byte{})

	command := exec.Command(c.cmd, c.params...)

	for key, value := range c.env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, value))
	}

	command.Stdout = io.MultiWriter(outb, os.Stdout)
	command.Stderr = io.MultiWriter(errb, os.Stderr)

	err := command.Run()
	command.Stdin = c.stdin

	return outb.Bytes(), errb.Bytes(), err
}
