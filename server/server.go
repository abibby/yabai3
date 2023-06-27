package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
)

type I3msgError struct {
	ParseError    bool   `json:"parse_error"`
	ErrorMessage  string `json:"error"`
	Input         string `json:"input"`
	ErrorPosition string `json:"errorposition"`
}

func (e *I3msgError) Error() string {
	if e.ParseError {
		return fmt.Sprintf("%s\n%s\n%s", e.Input, e.ErrorPosition, e.ErrorMessage)
	}
	return ""
}

type CommandResult struct {
	Success bool `json:"success"`
	*I3msgError
}

const PORT = 3141

func Serve(changeMode func(mode string) error) {
	http.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			sendError(w, err)
			return
		}

		results := []*CommandResult{}

		commands := badparser.SplitCommands(badparser.TokenizeLine(string(b)))
		for _, command := range commands {
			err := run.Command(command, changeMode)
			var msgErr *I3msgError
			if err != nil {
				msgErr = &I3msgError{
					ParseError:   true,
					ErrorMessage: err.Error(),
				}
			}
			results = append(results, &CommandResult{
				Success:    err == nil,
				I3msgError: msgErr,
			})

		}
		json.NewEncoder(w).Encode(results)
	})
	log.Print("server running")
	http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
}

func sendError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode([]*CommandResult{
		{
			Success: false,
			I3msgError: &I3msgError{
				ParseError:   true,
				ErrorMessage: err.Error(),
			},
		},
	})
}
