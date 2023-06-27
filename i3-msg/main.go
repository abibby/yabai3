package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/abibby/yabai3/server"
)

func main() {
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/command", server.PORT),
		"application/json",
		bytes.NewBufferString(strings.Join(os.Args[1:], " ")),
	)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode([]*server.CommandResult{
			{
				Success: false,
				I3msgError: &server.I3msgError{
					ParseError:   true,
					ErrorMessage: err.Error(),
				},
			},
		})
		os.Exit(1)
	}
	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)
}
