package goinput

import (
	"os/exec"
	"strings"
	"strconv"
)

func checkCmdExist(cmd string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		panic(err)
	}
}

// Although the commands can be used without receiver, I assign them to Mouse/Keyboard.

// Getmouselocation get mouse current location
func (*Mouse) Getmouselocation() [2]int {
	output, err := exec.Command("xdotool", "getmouselocation").Output()
	if err != nil {
		panic(err)
	}
	keyValues := strings.Split(string(output), " ")
	height, width := keyValues[1], keyValues[0]
	height = strings.Split(height, ":")[1]
	width = strings.Split(width, ":")[1]
	wI, _ := strconv.Atoi(width)
	hI, _ := strconv.Atoi(height)
	return [2]int{wI, hI}
}

// Mousemove move mouse to location x, y
func (*Mouse) Mousemove(x, y int) {
	exec.Command("xdotool", "mousemove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

// Keyup Release key on keyboard
func (*Keyboard) Keyup(key string) {
	exec.Command("xdotool", "keyup", key).Run()
}

// Keydown Press key on keyboard
func (*Keyboard) Keydown(key string) {
	exec.Command("xdotool", "keydown", key).Run()
}

// Mouseclick Click middle/right/left button
func (*Mouse) Mouseclick(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "click", code)
}

// Mouseup Release middle/right/left button
func (*Mouse) Mouseup(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mouseup", code)
}

// Mousedown Press middle/right/left button
func (*Mouse) Mousedown(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mousedown", code)
}
