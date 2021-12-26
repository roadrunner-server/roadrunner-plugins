package server

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (p *Plugin) runOnInitCommand() error {
	const op = errors.Op("server_on_init")
	stopCh := make(chan struct{}, 1)

	cmd := p.createProcess(p.cfg.OnInit.Env, p.cfg.OnInit.Command)
	timer := time.NewTimer(p.cfg.OnInit.ExecTimeout)

	err := cmd.Start()
	if err != nil {
		return errors.E(op, err)
	}

	go func() {
		errW := cmd.Wait()
		if errW != nil {
			p.log.Error("process wait", zap.Error(errW))
		}

		stopCh <- struct{}{}
	}()

	select {
	case <-timer.C:
		err = cmd.Process.Kill()
		if err != nil {
			p.log.Error("process killed", zap.Error(err))
		}
		return nil

	case <-stopCh:
		timer.Stop()
		return nil
	}
}

func (p *Plugin) Write(data []byte) (int, error) {
	p.log.Info(string(data))
	return len(data), nil
}

// create command for the process
func (p *Plugin) createProcess(env map[string]string, cmd string) *exec.Cmd {
	// cmdArgs contain command arguments if the command in form of: php <command> or ls <command> -i -b
	var cmdArgs []string
	var command *exec.Cmd
	cmdArgs = append(cmdArgs, strings.Split(cmd, " ")...)
	if len(cmdArgs) < 2 {
		command = exec.Command(cmd)
	} else {
		command = exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
	}

	// set env variables from the config
	if len(env) > 0 {
		for k, v := range env {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
	}

	// append system envs
	command.Env = append(command.Env, os.Environ()...)
	// redirect stderr and stdout into the Write function of the process.go
	command.Stderr = p
	command.Stdout = p

	return command
}
