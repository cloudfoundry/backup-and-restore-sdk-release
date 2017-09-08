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

	c := exec.Command(cmd, params...)
	for key, value := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", key, value))
	}
	c.Stdout = io.MultiWriter(outb, os.Stdout)
	c.Stderr = io.MultiWriter(errb, os.Stderr)
	err := c.Run()

	return outb.Bytes(), errb.Bytes(), err
}
