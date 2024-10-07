package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/abibby/yabai3/server"
	"github.com/spf13/pflag"
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
	msgType := pflag.StringP("type", "t", "command", "Send ipc message.")
	quiet := pflag.BoolP("quiet", "q", false, "Only send ipc message and suppress the output of the response.")
	monitor := pflag.BoolP("monitor", "m", false, "Only send ipc message and suppress the output of the response.")
	pflag.Parse()

	args := pflag.Args()
	if (*msgType) == "command" && len(args) == 0 {
		fmt.Println("i3-msg <command...>")
		return
	}
	conn, err := net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", server.PORT))
	check(err)

	defer conn.Close()

	err = json.NewEncoder(conn).Encode(&server.Request{
		Type:    *msgType,
		Message: strings.Join(args, " "),
		Monitor: *monitor,
	})
	check(err)

	if !*quiet {
		_, err = io.Copy(os.Stdout, conn)
		check(err)
	}
}
