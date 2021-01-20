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

func delegate(ctx context.Context, cli command.Cli, args ...string) {
	options := types.ImagePullOptions{}
	_, _, err := cli.Client().ImageInspectWithRaw(ctx, image)
	if err != nil {
		responseBody, err := cli.Client().ImagePull(ctx, image, options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		//nolint: errcheck
		defer responseBody.Close()
		err = jsonmessage.DisplayJSONMessagesStream(responseBody, bytes.NewBuffer(nil), os.Stdout.Fd(), false, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	hostConfig := container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/Dockerfile", dockerfilePath)},
	}

	result, err := cli.Client().ContainerCreate(ctx, &config, &hostConfig, nil, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	containerID := result.ID
	defer cli.Client().ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})

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
	cli.Client().ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Close()
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
			fmt.Fprintln(os.Stderr, "here?")
			os.Exit(1)
		}
	}

	os.Exit(0)
}
