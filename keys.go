package main

import (
	"log"
	"strings"

	"golang.design/x/hotkey"
)

var modMap = map[string]hotkey.Modifier{
	"ctrl":  hotkey.ModCtrl,
	"shift": hotkey.ModShift,
	"mod4":  hotkey.ModOption,
	"opt":   hotkey.ModOption,
	"mod1":  hotkey.ModCmd,
	"cmd":   hotkey.ModCmd,
}

var keyMap = map[string]hotkey.Key{
	"space": hotkey.KeySpace,
	"1":     hotkey.Key1,
	"2":     hotkey.Key2,
	"3":     hotkey.Key3,
	"4":     hotkey.Key4,
	"5":     hotkey.Key5,
	"6":     hotkey.Key6,
	"7":     hotkey.Key7,
	"8":     hotkey.Key8,
	"9":     hotkey.Key9,
	"0":     hotkey.Key0,
	"a":     hotkey.KeyA,
	"b":     hotkey.KeyB,
	"c":     hotkey.KeyC,
	"d":     hotkey.KeyD,
	"e":     hotkey.KeyE,
	"f":     hotkey.KeyF,
	"g":     hotkey.KeyG,
	"h":     hotkey.KeyH,
	"i":     hotkey.KeyI,
	"j":     hotkey.KeyJ,
	"k":     hotkey.KeyK,
	"l":     hotkey.KeyL,
	"m":     hotkey.KeyM,
	"n":     hotkey.KeyN,
	"o":     hotkey.KeyO,
	"p":     hotkey.KeyP,
	"q":     hotkey.KeyQ,
	"r":     hotkey.KeyR,
	"s":     hotkey.KeyS,
	"t":     hotkey.KeyT,
	"u":     hotkey.KeyU,
	"v":     hotkey.KeyV,
	"w":     hotkey.KeyW,
	"x":     hotkey.KeyX,
	"y":     hotkey.KeyY,
	"z":     hotkey.KeyZ,

	"return": hotkey.KeyReturn,
	"escape": hotkey.KeyEscape,
	"delete": hotkey.KeyDelete,
	"tab":    hotkey.KeyTab,

	"left":  hotkey.KeyLeft,
	"right": hotkey.KeyRight,
	"up":    hotkey.KeyUp,
	"down":  hotkey.KeyDown,

	"f1":  hotkey.KeyF1,
	"f2":  hotkey.KeyF2,
	"f3":  hotkey.KeyF3,
	"f4":  hotkey.KeyF4,
	"f5":  hotkey.KeyF5,
	"f6":  hotkey.KeyF6,
	"f7":  hotkey.KeyF7,
	"f8":  hotkey.KeyF8,
	"f9":  hotkey.KeyF9,
	"f10": hotkey.KeyF10,
	"f11": hotkey.KeyF11,
	"f12": hotkey.KeyF12,
	"f13": hotkey.KeyF13,
	"f14": hotkey.KeyF14,
	"f15": hotkey.KeyF15,
	"f16": hotkey.KeyF16,
	"f17": hotkey.KeyF17,
	"f18": hotkey.KeyF18,
	"f19": hotkey.KeyF19,
	"f20": hotkey.KeyF20,

	"comma":  43,
	"period": 47,

	"function": 0x3F,
}

func keys(keysStr string) ([]hotkey.Modifier, hotkey.Key) {
	mods := []hotkey.Modifier{}
	key := hotkey.Key(0)

	keys := strings.Split(keysStr, "+")
	for _, k := range keys {
		if mod, ok := modMap[strings.ToLower(k)]; ok {
			mods = append(mods, mod)
		}
		if ke, ok := keyMap[strings.ToLower(k)]; ok {
			key = ke
		}
	}

	if len(mods) == 1 && mods[0] == hotkey.ModCtrl && (key == hotkey.KeyC || key == hotkey.KeyD) {
		log.Fatal("cannot use ctrl+c or ctrl+d as a hotkey")
	}
	return mods, key
}
