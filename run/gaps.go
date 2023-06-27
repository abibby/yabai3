package run

import (
	"fmt"

	"github.com/abibby/yabai3/yabai"
)

func SetGaps(inner, outer int) error {
	err := yabai.Yabai("config", "window_gap", fmt.Sprint(inner))
	if err != nil {
		return err
	}

	err = yabai.Yabai("config", "top_padding", fmt.Sprint(outer))
	if err != nil {
		return err
	}
	err = yabai.Yabai("config", "bottom_padding", fmt.Sprint(outer))
	if err != nil {
		return err
	}
	err = yabai.Yabai("config", "left_padding", fmt.Sprint(outer))
	if err != nil {
		return err
	}
	err = yabai.Yabai("config", "right_padding", fmt.Sprint(outer))
	if err != nil {
		return err
	}
	return nil
}
