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

// Getlocation get mouse current location.
func (*Mouse) Getlocation() [2]int {
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

// Move move mouse to location x, y.
func (*Mouse) Move(x, y int) {
	exec.Command("xdotool", "mousemove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

// Downup Press and release key on keyboard.
// Check https://www.linux.org/threads/xdotool-keyboard.10528/ for supported keysequence.
func (*Keyboard) Downup(key string) {
	exec.Command("xdotool", "key", key).Run()
}

// Up Release key on keyboard.
// Check https://www.linux.org/threads/xdotool-keyboard.10528/ for supported keysequence.
func (*Keyboard) Up(key string) {
	exec.Command("xdotool", "keyup", key).Run()
}

// Down Press key on keyboard.
// Check https://www.linux.org/threads/xdotool-keyboard.10528/ for supported keysequence.
func (*Keyboard) Down(key string) {
	exec.Command("xdotool", "keydown", key).Run()
}

// Click Click middle/right/left button.
func (*Mouse) Click(key string) {
	code := xdotoolMouseMap[strings.ToUpper(key)]
	exec.Command("xdotool", "click", code)
}

// Up Release middle/right/left button.
func (*Mouse) Up(key string) {
	code := xdotoolMouseMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mouseup", code)
}

// Down Press middle/right/left button.
func (*Mouse) Down(key string) {
	code := xdotoolMouseMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mousedown", code)
}
