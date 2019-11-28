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

func Getmouselocation() [2]int {
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

func Mousemove(x, y int) {
	exec.Command("xdotool", "mousemove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

func Keyup(key string) {
	exec.Command("xdotool", "keyup", key).Run()
}

func Keydown(key string) {
	exec.Command("xdotool", "keydown", key).Run()
}

func Mouseclick(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "click", code)
}

func Mouseup(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mouseup", code)
}

func Mousedown(key string) {
	code := xdotoolMap[strings.ToUpper(key)]
	exec.Command("xdotool", "mousedown", code)
}
