package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/abibby/yabai3/server"
)

func check(err error) {
	if err == nil {
		return
	}
	jsonErr := json.NewEncoder(os.Stdout).Encode([]*server.CommandResult{
		{
			Success: false,
			I3msgError: &server.I3msgError{
				ParseError:   true,
				ErrorMessage: err.Error(),
			},
		},
	})
	if jsonErr != nil {
		panic(errors.Join(jsonErr, err))
	}
	os.Exit(1)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("i3-msg <command...>")
		return
	}
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/command", server.PORT),
		"application/json",
		bytes.NewBufferString(strings.Join(args, " ")),
	)
	check(err)
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	check(err)
}
