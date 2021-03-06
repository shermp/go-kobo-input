/*
	go-kobo-input - Basic touch input handling for Kobo ereaders
    Copyright (C) 2018 Sherman Perry

    This file is part of go-kobo-input.

    go-kobo-input is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    go-kobo-input is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with go-kobo-input.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	// Get the input event filepath
	touchPath := "/dev/input/event1"
	// Create an input object
	t := koboin.New(touchPath, 1080, 1440)
	if t == nil {
		return
	}
	defer t.Close()

	// Create and init FBInk
	cfg := gofbink.FBInkConfig{}
	rCfg := gofbink.RestrictedConfig{}
	rCfg.Fontmult = 3
	rCfg.Fontname = gofbink.IBM
	fb := gofbink.New(&cfg, &rCfg)
	fb.Open()
	defer fb.Close()
	fb.Init(&cfg)

	// Use kobo-sim-usb to enter USBMS mode where we can use the
	// touchscreen without unintended presses in Nickel
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

	// Load a keymap file for the OSK
	keymapJSON, _ := ioutil.ReadFile("./keymap-en_us.json")
	km := osk.KeyMap{}
	json.Unmarshal(keymapJSON, &km)
	// Create an OSK
	vk, _ := osk.New(&km, 1080, 1440)

	// Generate an image of the OSK
	vkPNG := "./osk-en_us.png"
	vkFont := "./Roboto-Medium.ttf"
	vk.CreateIMG(vkPNG, vkFont)
	// Print the image to the screen. Its position on screen should match that stored
	// in the keyboard object
	fb.PrintImage(vkPNG, int16(vk.StartCoords.X), int16(vk.StartCoords.Y), &cfg)
	runeStr := []rune{}
	upperCase := false
	cfg.Row = 16
	// Read the input from the touch screen
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
