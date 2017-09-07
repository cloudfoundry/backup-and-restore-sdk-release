package runner

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Run(cmd string, params []string, env map[string]string) ([]byte, []byte, error) {
	outb := bytes.NewBuffer([]byte{})
	errb := bytes.NewBuffer([]byte{})

	c := exec.Command(cmd, params...)
	for key, value := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", key, value))
	}
	c.Stdout = outb
	c.Stderr = errb
	err := c.Run()

	return outb.Bytes(), errb.Bytes(), err
}
