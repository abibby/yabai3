package bar

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
)

func StartCommand(ctx context.Context, command string) (io.WriteCloser, io.ReadCloser, *exec.Cmd) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	cmd.Stderr = NewPrefixWriter(os.Stderr, "bar | ")
	// cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start status command: %v", err)
	}

	return stdin, stdout, cmd
}
