package main

import (
	"context"
	"fmt"

	"github.com/abibby/yabai3/bar"
)

func main() {
	stdin, stdout, cmd := bar.StartCommand(context.Background(), "/Users/abibby/go/bin/i3gobar")
	defer stdin.Close()
	defer stdout.Close()
	updates := bar.Process(stdout)

	go func() {
		for b := range updates {
			fmt.Println(b.String())
		}
	}()
	cmd.Wait()
}
