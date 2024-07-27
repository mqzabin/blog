package extensions

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

type Command struct {
	command string
	args    []string
	envs    map[string]string
}

func NewCommand(command string, args ...string) *Command {
	return &Command{
		command: command,
		args:    args,
	}
}

func (c *Command) Args(args ...string) *Command {
	c.args = append(c.args, args...)

	return c
}

func (c *Command) Env(name, value string) *Command {
	if c.envs == nil {
		c.envs = make(map[string]string, 1)
	}

	c.envs[name] = value

	return c
}

func (c *Command) Run(ctx context.Context) error {
	successWriter := NewColorWriter(color.New(color.FgGreen), os.Stdout)
	errorWriter := NewColorWriter(color.New(color.FgRed), os.Stdout)

	cmd := exec.CommandContext(ctx, c.command, c.args...)
	cmd.Stdout = successWriter
	cmd.Stderr = errorWriter

	for k, v := range c.envs {
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%s", k, v))
	}

	if err := cmd.Run(); err != nil {
		errorWriter.Write([]byte("failed to run command:\n" + err.Error() + "\n"))
	}

	return nil
}
