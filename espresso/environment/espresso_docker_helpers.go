package environment

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// This is a reliable way to determine if we are running on Linux as a runtime
// check.
var isRunningOnLinux = runtime.GOOS == "linux"

// DockerContainerInfo is a struct that contains information about a Docker
// Container that was launched by the DockerCli struct.
// This is an informational snapshot only, and is not guaranteed to represent
// the current state of the container.
type DockerContainerInfo struct {
	// The container ID of the Docker container that is running the
	// Espresso Dev Node.
	// This is useful for further interaction with docker concerning
	// the specific Dev Node
	ContainerID string

	// The Port Map of the Resulting Docker Container
	PortMap map[string][]string
}

// DockerContainerConfig is a configuration struct that is used to configure
// the launching of a Docker Container
type DockerContainerConfig struct {
	Image string

	Environment map[string]string

	Ports []string

	Network  string
	AutoRM   bool
	Platform string
	Name     string
}

// DockerBuildArg is a configuration struct that is used to pass
// 'ARG' parameters when building a Docker Image
type DockerBuildArg struct {
	Name  string
	Value string
}

// DockerCli is a simple implementation of a Docker Client that is used to
// launch Docker Containers
type DockerCli struct{}

// LaunchContainer launches a Docker Container with the given configuration
// and returns the resulting Docker Container Info
//
// The Container will automatically be stopped when the given context is
// completed. This is done by spawning a goroutine that is blocked by the
// context that is passed in's Done channel.
func (d *DockerCli) LaunchContainer(ctx context.Context, config DockerContainerConfig) (DockerContainerInfo, error) {
	originalContext := ctx

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Remove existing container with the same name if it exists
	if config.Name != "" {
		// Try to remove the container, ignore errors if it doesn't exist
		removeCmd := exec.CommandContext(ctx, "docker", "rm", "-f", config.Name)
		_ = removeCmd.Run() // Ignore errors - container might not exist
	}

	outputBuffer := new(bytes.Buffer)
	var args []string
	// Let's build the arguments for the docker launch command
	{

		args = append(args, "run", "-d")

		if config.AutoRM {
			args = append(args, "--rm")
		}

		if config.Network != "" {
			args = append(args, "--network", config.Network)
		}

		if config.Network != "host" {
			for _, port := range config.Ports {
				args = append(args, "-p", port)
			}
		}
		// Add platform support
		if config.Platform != "" {
			args = append(args, "--platform", config.Platform)
		}

		if config.Name != "" {
			args = append(args, "--name", config.Name)
		}

		for key, value := range config.Environment {
			args = append(args, "-e", key+"="+value)
		}

		args = append(args, config.Image)
	}

	var containerID string
	{
		launchContainerCmd := exec.CommandContext(
			ctx,
			"docker",
			args...,
		)

		// A buffer to collect the output of the command, so we can retrieve the
		// Container ID.
		launchContainerCmd.Stdout = outputBuffer

		stderrBuffer := new(bytes.Buffer)
		launchContainerCmd.Stderr = stderrBuffer
		launchContainerCmd.Stdout = outputBuffer

		if err := launchContainerCmd.Run(); err != nil {
			return DockerContainerInfo{}, fmt.Errorf("failed to launch docker container: %w\nstderr: %s", err, stderrBuffer.String())
		}

		containerID = strings.TrimSpace(outputBuffer.String())
	}

	// Let's setup a cleanup function to stop the container, should we
	// need to.

	stopContainer := func() error {
		return d.StopContainer(context.Background(), containerID)
	}

	// We spin up a goroutine that will clean us up when the original context
	// dies
	go (func(ctx context.Context) {
		// Wait for the context that governs us to tell us to die
		<-ctx.Done()

		err := stopContainer()
		if err != nil {
			log.Printf("failed to stop docker container: %v", err)
		}
	})(originalContext)

	// We have the container ID.  Let's get our Ports

	portMap := map[string][]string{}
	containerInfo := DockerContainerInfo{ContainerID: containerID, PortMap: portMap}
	if config.Network == "host" {
		// If we're running on the host network, we don't need to do anything
		// special to get the ports.  They are the same as the ones we specified
		// in the config.

		for _, port := range config.Ports {
			portMap[port] = []string{
				fmt.Sprintf("0.0.0.0:%s", port),
			}
		}
	} else {
		for _, portToFind := range config.Ports {
			outputBuffer.Reset()
			// Let's find out what our assigned ports ended up being
			determinePortCmd := exec.CommandContext(
				ctx,
				"docker",
				"port",
				containerID,
				portToFind,
			)
			determinePortCmd.Stdout = outputBuffer

			if err := determinePortCmd.Run(); err != nil {
				return containerInfo, err
			}

			lineReader := bufio.NewReader(outputBuffer)

			for {
				line, _, err := lineReader.ReadLine()
				if err == io.EOF {
					// we consumed all of it
					break
				}

				if err != nil {
					return DockerContainerInfo{ContainerID: containerID}, err
				}

				if len(line) == 0 {
					// empty line, ignore
					continue
				}

				portMap[portToFind] = append(portMap[portToFind], string(line))
			}
		}
	}

	return containerInfo, nil
}

// DockerInspectContainerStateHealth is a struct that contains information
// about the health of a Docker Container.
// This struct is created based on the observed output of the `docker inspect`
// command.  It is not complete, and is not guaranteed to be correct.

type DockerInspectContainerStateHealth struct {
	Status        string
	FailingStreak uint
	// Log

}

// DockerInspectContainerState is a struct that contains information
// about the state of a Docker Container.
// This struct is created based on the observed output of the `docker inspect`
// command.  It is not complete, and is not guaranteed to be correct.
type DockerInspectContainerState struct {
	Status     string
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        uint
	ExitCode   uint
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
	Health     DockerInspectContainerStateHealth
}

// DockerInspectContainerInfo is a struct that contains information
// about a Docker Container.
// This is an informational snapshot only, and is not guaranteed to represent
// the current state of the container.
type DockerInspectContainerInfo struct {
	Id              string
	Created         time.Time
	Path            string
	Args            []string
	State           DockerInspectContainerState
	Image           string
	ResolveConfPath string
	HostnamePath    string
	HostsPath       string
	LogPath         string
	Name            string
	RestartCount    uint
	Driver          string
	Platform        string
	MountLabel      string
	ProcessLabel    string
	AppArmorProfile string
	// ExecIds []string
}

// ErrDockerInspectRequiresAtLeastOneContainerID is an error that indicates
// that in order to run cocker inspect, we need to specify container IDs
// to inspect.  We can specify multiple, but at least one is required.
var ErrDockerInspectRequiresAtLeastOneContainerID = errors.New("docker inspect requires at least one container ID")

// Inspect runs the `docker inspect` command with the given containerIDs, and
// returns the given parsed output from the json representation of the command
func (d *DockerCli) Inspect(ctx context.Context, containerIDs ...string) ([]DockerInspectContainerInfo, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(containerIDs) < 1 {
		return nil, ErrDockerInspectRequiresAtLeastOneContainerID
	}

	outputBuffer := new(bytes.Buffer)

	args := make([]string, 0, len(containerIDs)+3)
	args = append(args, "inspect", "--format", "json")
	args = append(args, containerIDs...)

	inspectCmd := exec.CommandContext(
		ctx,
		"docker",
		args...,
	)

	inspectCmd.Stdout = outputBuffer

	if err := inspectCmd.Run(); err != nil {
		return nil, err
	}

	var result []DockerInspectContainerInfo
	err := json.NewDecoder(outputBuffer).Decode(&result)
	return result, err
}

// InspectOne is a specialized case of DockerCli.Inspect that only runs on a
// single containerID
func (d *DockerCli) InspectOne(ctx context.Context, containerID string) (DockerInspectContainerInfo, error) {
	containerInfos, err := d.Inspect(ctx, containerID)

	if len(containerInfos) <= 0 {
		return DockerInspectContainerInfo{}, errors.New("no results")
	}

	return containerInfos[0], err
}

// DockerContainerNotRunningError is an error that indicates that a Docker
// Container is not running.
type DockerContainerNotRunningError struct {
	ContainerID string
}

// Error implements error
func (e DockerContainerNotRunningError) Error() string {
	return fmt.Sprintf("unable to stop container %s, it is not running", e.ContainerID)
}

// StopContainer stops a Docker Container with the given container ID
func (d *DockerCli) StopContainer(ctx context.Context, containerID string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result, err := d.InspectOne(ctx, containerID)
	if err != nil {
		return err
	}

	if !result.State.Running {
		return DockerContainerNotRunningError{containerID}
	}

	stopCmd := exec.CommandContext(
		ctx,
		"docker",
		"stop",
		containerID,
	)

	return stopCmd.Run()
}

// Logs retrieves the logs from a Docker Container with the given
// container ID
//
// This command will keep running until the passed context is cancelled.
func (d *DockerCli) Logs(ctx context.Context, containerID string) (io.Reader, error) {
	logsCmd := exec.CommandContext(
		ctx,
		"docker",
		"logs",
		"-f",
		containerID,
	)
	reader, err := logsCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := logsCmd.Start(); err != nil {
		return nil, err
	}

	// This needs to be launched in the background
	go func(cmd *exec.Cmd) {
		// Wait for the context to be cancelled
		<-ctx.Done()

		// We don't really have a great opportunity to inspect any error
		// returned by this command
		err = cmd.Wait()
	}(logsCmd)

	return reader, err
}

// Build builds a Docker Image with the given tag, dockerfile, target, context, and build arguments.
func (d *DockerCli) Build(ctx context.Context, tag string, dockerfile string, target string, context string, buildArgs ...DockerBuildArg) error {
	args := []string{
		"build",
		"--tag",
		tag,
		"--file",
		dockerfile,
		"--target",
		target,
	}
	for _, arg := range buildArgs {
		args = append(args, "--build-arg", arg.Name+"="+arg.Value)
	}
	args = append(args, context)

	build := exec.CommandContext(ctx, "docker", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	return build.Run()
}
