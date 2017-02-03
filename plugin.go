package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

const (
	maxRetry   = 15
	dockerBin  = "/usr/local/bin/docker"
	dockerdBin = "/usr/local/bin/dockerd"
)

// Plugin defines the GCR plugin parameters.
type Plugin struct {
	// Docker push is skipped.
	DryRun bool `envconfig:"DRY_RUN"`
	// Docker lounch debug enabling.
	Debug bool `envconfig:"DEBUG"`

	// GCR registry address.
	Registry string `envconfig:"REGISTRY" default:"gcr.io"`
	// GCR authorization key.
	AuthKey string `envconfig:"AUTH_KEY" required:"true"`
	// Docker daemon storage driver.
	StorageDriver string `envconfig:"STORAGE_DRIVER"`

	// Docker build using default named tag.
	Name string `envconfig:"DRONE_COMMIT_SHA" default:"00000000"`
	// Docker build repository.
	Repo string `envconfig:"REPO" required:"true"`
	// Docker build Dockerfile.
	Dockerfile string `envconfig:"DOCKERFILE" default:"Dockerfile"`
	// Docker build context.
	Context string `envconfig:"CONTEXT" default:"."`
	// Docker build tags.
	Tags []string `envconfig:"TAGS" default:"latest"`
	// Docker build args.
	Args []string `envconfig:"ARGS"`
}

// Exec executes the plugin step.
func (p Plugin) Exec() error {
	cmd := p.cmdDaemon()
	if p.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}
	go func() {
		trace(cmd)
		cmd.Run()
	}()

	// Poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < maxRetry; i++ {
		cmd := commandInfo()
		if err := cmd.Run(); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}

	cmd = p.cmdLogin()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error authenticating: %s", err)
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion()) // docker version
	cmds = append(cmds, commandInfo())    // docker info
	cmds = append(cmds, p.cmdBuild())     // docker build

	for _, tag := range p.Tags {
		cmds = append(cmds, p.cmdTag(tag)) // docker tag

		if p.DryRun == false {
			cmds = append(cmds, p.cmdPush(tag)) // docker push
		}
	}

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		trace(cmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Plugin) cmdLogin() *exec.Cmd {
	args := []string{
		"login",
		"-u", "_json_key",
		"-p", p.AuthKey,
		p.Registry,
	}

	return exec.Command(dockerBin, args...)
}

// helper function to create the docker build command.
func (p Plugin) cmdBuild() *exec.Cmd {
	args := []string{
		"build",
		"--pull=true",
		"--rm=true",
		"-t", p.Name,
		"-f", p.Dockerfile,
	}

	args = append(args, p.Context)
	for _, arg := range p.Args {
		args = append(args, "--build-arg", arg)
	}

	return exec.Command(dockerBin, args...)
}

// helper function to create the docker tag command.
func (p Plugin) cmdTag(tag string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", p.Repo, tag)
	return exec.Command(
		dockerBin, "tag", p.Name, target,
	)
}

// helper function to create the docker push command.
func (p Plugin) cmdPush(tag string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", p.Repo, tag)
	return exec.Command(dockerBin, "push", target)
}

// helper function to create the docker daemon command.
func (p Plugin) cmdDaemon() *exec.Cmd {
	var args []string
	if p.StorageDriver != "" {
		args = append(args, "-s", p.StorageDriver)
	}

	return exec.Command(dockerdBin, args...)
}

// helper function to create the docker info command.
func commandVersion() *exec.Cmd {
	return exec.Command(dockerBin, "version")
}

// helper function to create the docker info command.
func commandInfo() *exec.Cmd {
	return exec.Command(dockerBin, "info")
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Println(cmd.Args)
	// logrus.WithField("args", cmd.Args).Debug("debug")
}
