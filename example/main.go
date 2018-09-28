package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"unicode"

	"github.com/shermp/go-fbink-v2/gofbink"
	"github.com/shermp/go-kobo-input/koboin"
	"github.com/shermp/go-osk/osk"
	"github.com/shermp/kobo-sim-usb/simusb"
)

func main() {
	touchPath := "/dev/input/event1"
	t := koboin.New(touchPath, 1080, 1440)
	if t == nil {
		return
	}
	defer t.Close()

	cfg := gofbink.FBInkConfig{}
	rCfg := gofbink.RestrictedConfig{}
	rCfg.Fontmult = 3
	rCfg.Fontname = gofbink.IBM
	fb := gofbink.New(&cfg, &rCfg)
	fb.Open()
	defer fb.Close()
	fb.Init(&cfg)

	u, err := simusb.New(fb)
	if err != nil {
		fmt.Println(err)
	}
	err = u.Start(true, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer u.End(true)

	fb.Println("Welcome to the test!")
	fb.Println("Have Fun!")

	keymapJSON, _ := ioutil.ReadFile("./keymap-en_us.json")
	km := osk.KeyMap{}
	json.Unmarshal(keymapJSON, &km)
	vk, _ := osk.New(&km, 1080, 1440)

	vkPNG := "./osk-en_us.png"
	vkFont := "./Roboto-Medium.ttf"
	vk.CreateIMG(vkPNG, vkFont)
	fb.PrintImage(vkPNG, int16(vk.StartCoords.X), int16(vk.StartCoords.Y), &cfg)
	runeStr := []rune{}
	upperCase := false
	cfg.Row = 16
L:
	for {
		x, y, err := t.GetInput()
		if err != nil {
			continue
		}
		k, err := vk.GetPressedKey(x, y)
		if err != nil {
			continue
		}
		if !k.IsKey {
			continue
		}
		switch k.KeyType {
		case osk.KTstandardChar:
			var key rune
			if upperCase {
				key = unicode.ToUpper(k.KeyCode)
			} else {
				key = unicode.ToLower(k.KeyCode)
			}
			runeStr = append(runeStr, key)
			fb.FBprint(string(runeStr), &cfg)
		case osk.KTbackspace:
			if len(runeStr) > 0 {
				runeStr[len(runeStr)-1] = 32
				runeStr = runeStr[:len(runeStr)-1]
				fb.FBprint(string(runeStr), &cfg)
			}
		case osk.KTcapsLock:
			upperCase = !upperCase
		case osk.KTcarriageReturn:
			break L
		default:
			continue
		}
	}
}
