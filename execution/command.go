package execution

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Command struct {
	cmd *exec.Cmd
}

const (
	COMMAND_STILL_RUNNING       = -1
	COMMAND_INTERRUPT_EXIT_CODE = 130
	COMMAND_UNKNOWN_EXIT_CODE   = 70
)

var (
	aliveSignal = syscall.Signal(0)
)

func (c *Command) String() string {
	return fmt.Sprintf("<Command(command='%v' (%v))>",
		c.cmd.Args,
		c.cmd)
}

func (c *Command) Stopped() bool {
	return c.cmd.ProcessState != nil
}

func (c *Command) ExitCode() int {
	if !c.Stopped() {
		log.WithField("command", c).Warn("Command is still running!")
		return COMMAND_STILL_RUNNING
	}

	waitStatus, ok := c.cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		log.Fatal("Cannot convert ProcessState to WaitStatus!")
		return COMMAND_UNKNOWN_EXIT_CODE
	}

	exitCode := waitStatus.ExitStatus()
	if exitCode < 0 {
		exitCode = COMMAND_INTERRUPT_EXIT_CODE
	}

	return exitCode
}

func (c *Command) Start() {
	go c.cmd.Run()
}

func (c *Command) Stop(signal os.Signal, timeout time.Duration) {
	if c.Stopped() {
		return
	}

	gracefulTimerChannel := time.Tick(2 * time.Millisecond)
	killTimerChannel := time.After(timeout)

	c.cmd.Process.Signal(signal)

	for {
		select {
		case <-gracefulTimerChannel:
			if c.Stopped() {
				break
			}
		case <-killTimerChannel:
			if c.Stopped() {
				break
			} else {
				log.Info("Graceful timeout expired, send kill signal")
				c.cmd.Process.Signal(syscall.SIGKILL)
			}
		}
	}
}

func NewCommand(commandToExecute []string, hasTTY bool) (cmd *Command) {
	cmd = &Command{
		cmd: exec.Command(commandToExecute[0], commandToExecute[1:]...),
	}

	return
}
