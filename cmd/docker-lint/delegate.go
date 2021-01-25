package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"os"
	"path/filepath"
)

var (
	ImageDigest = "sha256:2021ffa2e860757aa28367dfe6b77cda95dfa3276781b3932130ef80a46612c5"
	image       = fmt.Sprintf("hadolint/hadolint@%s", ImageDigest)
)

type removeContainerFunc func() error

func delegate(ctx context.Context, cli command.Cli, args ...string) {
	err := pullImage(ctx, cli)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	containerID, removeContainer, err := createContainer(ctx, cli, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	//nolint: errcheck
	defer removeContainer()

	streamFunc, err := startContainer(ctx, cli, containerID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer streamFunc()

	statusc, errc := cli.Client().ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errc:
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case s := <-statusc:
		switch s.StatusCode {
		case 0:
		default:
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func removeContainer(ctx context.Context, cli command.Cli, containerID string) removeContainerFunc {
	return func() error {
		return cli.Client().ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	}
}

func pullImage(ctx context.Context, cli command.Cli) error {
	options := types.ImagePullOptions{}
	_, _, err := cli.Client().ImageInspectWithRaw(ctx, image)
	if err != nil {
		responseBody, err := cli.Client().ImagePull(ctx, image, options)
		if err != nil {
			return err
		}
		//nolint: errcheck
		defer responseBody.Close()
		return jsonmessage.DisplayJSONMessagesStream(responseBody, bytes.NewBuffer(nil), os.Stdout.Fd(), false, nil)

	}
	return nil
}

func createContainer(ctx context.Context, cli command.Cli, args ...string) (string, removeContainerFunc, error) {
	cmdArgs := append(strslice.StrSlice{"hadolint"}, args...)

	config := container.Config{
		AttachStdout:    true,
		AttachStderr:    true,
		Image:           image,
		Cmd:             cmdArgs,
		NetworkDisabled: true,
	}
	dockerfilePath, err := filepath.Abs(args[len(args)-1])
	if err != nil {
		return "", nil, err
	}
	hostConfig := container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/Dockerfile", dockerfilePath)},
	}

	result, err := cli.Client().ContainerCreate(ctx, &config, &hostConfig, nil, "")
	if err != nil {
		return "", nil, err
	}
	removeContainerFunc := removeContainer(ctx, cli, result.ID)
	return result.ID, removeContainerFunc, nil
}

func startContainer(ctx context.Context, cli command.Cli, containerID string) (func(), error) {
	resp, err := cli.Client().ContainerAttach(ctx, containerID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	go func() {
		for {
			_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
		}
	}()
	return resp.Close, cli.Client().ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}