package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/abibby/salusa/set"
	"github.com/abibby/yabai3/badparser"
	"github.com/abibby/yabai3/run"
	"github.com/abibby/yabai3/yabai"
	"github.com/davecgh/go-spew/spew"
)

// https://man.archlinux.org/man/extra/i3-wm/i3-msg.1.en

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

type I3MsgServer struct {
	listener   net.Listener
	changeMode func(mode string) error
	restart    func() error

	modeChangeEventsMtx *sync.Mutex
	modeChangeEvents    set.Set[chan string]
}

func New() *I3MsgServer {
	return &I3MsgServer{
		modeChangeEventsMtx: &sync.Mutex{},
		modeChangeEvents:    set.Set[chan string]{},
	}
}

func (s *I3MsgServer) Start(ctx context.Context, changeMode func(mode string) error, restart func() error) error {

	s.changeMode = changeMode
	s.restart = restart

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", PORT))
	if err != nil {
		return err
	}
	s.listener = l

	go func() {
		defer l.Close()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for {
			c, err := l.Accept()
			if err != nil {
				log.Printf("i3-msg server: accept: %v", err)
				return
			}
			go s.rootHandler(ctx, c)
		}
	}()
	return nil
}

type Request struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Monitor bool   `json:"monitor"`
	ctx     context.Context
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (s *I3MsgServer) rootHandler(ctx context.Context, c net.Conn) {
	defer func() {
		spew.Dump(c)
		c.Close()
	}()

	r := json.NewDecoder(c)
	w := json.NewEncoder(c)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		req := &Request{}
		err := r.Decode(req)
		if errors.Is(err, net.ErrClosed) {
			return
		} else if err != nil {
			log.Printf("i3-msg server: handle: %v", err)
			return
		}

		go func() {
			req.ctx = ctx

			err = s.processRequest(w, req)
			if err != nil {
				log.Printf("i3-msg server: handle: %v", err)
				return
			}
			if !req.Monitor {
				c.Close()
				return
			}
		}()
	}
}

func (s *I3MsgServer) processRequest(w *json.Encoder, r *Request) error {
	switch r.Type {
	case "command":
		return s.command(w, r)
	case "get_workspaces":
		return s.getWorkspaces(w, r)
	case "subscribe":
		return s.subscribe(w, r)
	default:
		return fmt.Errorf("i3-msg server: handle: invalid type: %s", r.Type)
	}
}

func (s *I3MsgServer) Close() error {
	errs := []error{}
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	s.listener = nil
	s.restart = nil
	s.changeMode = nil
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *I3MsgServer) ModeChanged(mode string) {
	s.modeChangeEventsMtx.Lock()
	defer s.modeChangeEventsMtx.Unlock()
	for e := range s.modeChangeEvents {
		e <- mode
	}
}

func (s *I3MsgServer) removeModeChangeEvents(events chan string) {
	s.modeChangeEventsMtx.Lock()
	defer s.modeChangeEventsMtx.Unlock()
	s.modeChangeEvents.Delete(events)
}

func (s *I3MsgServer) addModeChangeEvents(events chan string) {
	s.modeChangeEventsMtx.Lock()
	defer s.modeChangeEventsMtx.Unlock()
	s.modeChangeEvents.Add(events)
}

func sendError(w http.ResponseWriter, err error) {
	jsonErr := json.NewEncoder(w).Encode([]*CommandResult{
		{
			Success: false,
			I3msgError: &I3msgError{
				ParseError:   true,
				ErrorMessage: err.Error(),
			},
		},
	})
	if jsonErr != nil {
		log.Printf("i3-msg server: %v", jsonErr)
	}
}

func (s *I3MsgServer) command(w *json.Encoder, r *Request) error {
	results := []*CommandResult{}

	commands := badparser.SplitCommands(badparser.TokenizeLine(r.Message))
	for _, command := range commands {
		err := run.Command(command, s.changeMode, s.restart)
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
	return w.Encode(results)
}

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Workspace struct {
	ID      int64  `json:"id"`
	Num     int    `json:"num"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Focused bool   `json:"focused"`
	Rect    Rect   `json:"rect"`
	Output  string `json:"output"`
	Urgent  bool   `json:"urgent"`
}

func (s *I3MsgServer) getWorkspaces(w *json.Encoder, r *Request) error {
	spaces, err := yabai.QuerySpaces()
	if err != nil {
		// sendError(w, err)
		return err
	}
	displays, err := yabai.QueryDisplays()
	if err != nil {
		// sendError(w, err)
		return err
	}

	workspaces := make([]*Workspace, len(spaces))
	for i, s := range spaces {
		var display *yabai.Display
		for _, d := range displays {
			if d.ID == s.DisplayIndex {
				display = d
				break
			}
		}
		workspaces[i] = &Workspace{
			ID:      int64(s.ID),
			Num:     s.Index,
			Name:    s.Label,
			Visible: s.IsVisible,
			Focused: s.HasFocus,
			Rect: Rect{
				X:      int(display.Frame.X),
				Y:      int(display.Frame.Y),
				Width:  int(display.Frame.Width),
				Height: int(display.Frame.Height),
			},
			Output: fmt.Sprint(s.DisplayIndex),
			Urgent: false,
		}
	}

	return w.Encode(workspaces)
}

func (s *I3MsgServer) subscribe(w *json.Encoder, r *Request) error {
	events := []string{}

	err := json.Unmarshal([]byte(r.Message), &events)
	if err != nil {
		return fmt.Errorf("i3-msg: subscribe: %w", err)
	}

	modeChanges := make(chan string)
	for _, e := range events {
		switch e {
		case "mode":
			s.addModeChangeEvents(modeChanges)
			defer s.removeModeChangeEvents(modeChanges)
		}
	}

	for {
		log.Print("loop")
		select {
		case <-r.Context().Done():
			return nil
		case mode := <-modeChanges:
			log.Print("mode ", mode)
			err := w.Encode(map[string]any{
				"change":       mode,
				"pango_markup": false,
			})
			if err != nil {
				return err
			}
		}
	}
}

// "get_workspaces": // Gets the current workspaces. The reply will be a JSON-encoded list of workspaces.
// "get_outputs": // Gets the current outputs. The reply will be a JSON-encoded list of outputs (see the reply section of docs/ipc, e.g. at https://i3wm.org/docs/ipc.html#_receiving_replies_from_i3).
// "get_tree": // Gets the layout tree. i3 uses a tree as data structure which includes every container. The reply will be the JSON-encoded tree.
// "get_marks": // Gets a list of marks (identifiers for containers to easily jump to them later). The reply will be a JSON-encoded list of window marks.
// "get_bar_config": // Gets the configuration (as JSON map) of the workspace bar with the given ID. If no ID is provided, an array with all configured bar IDs is returned instead.
// "get_binding_modes": // Gets a list of configured binding modes.
// "get_version": // Gets the version of i3. The reply will be a JSON-encoded dictionary with the major, minor, patch and human-readable version.
// "get_config": // Gets the currently loaded i3 configuration.
// "send_tick": // Sends a tick to all IPC connections which subscribe to tick events.
// "subscribe": // The payload of the message describes the events to subscribe to. Upon reception, each event will be dumped as a JSON-encoded object. See the -m option for continuous monitoring.
