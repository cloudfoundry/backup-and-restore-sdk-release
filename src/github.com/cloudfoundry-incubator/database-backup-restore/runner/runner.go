package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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
	return Command{cmd: c.cmd, params: append(c.params, params...), env: c.env, stdin: c.stdin}
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

	command.Env = c.buildEnvStrings()

	command.Stdout = io.MultiWriter(outb, os.Stdout)
	command.Stderr = io.MultiWriter(errb, os.Stderr)
	command.Stdin = c.stdin

	err := command.Run()

	return outb.Bytes(), errb.Bytes(), err
}

func (c Command) buildEnvStrings() []string {
	var env []string
	for key, value := range c.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

func (c Command) String() string {
	env := strings.Join(c.buildEnvStrings(), " ")
	params := strings.Join(c.params, " ")
	return fmt.Sprintf("%s %s %s", env, c.cmd, params)
}
