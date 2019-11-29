package main

import (
	"fmt"
	"github.com/kiki0805/goinput"
	"time"
)

func main() {
	fmt.Println("Testing listener...")
	mouse := goinput.NewMouse()
	mouse.Listen()
	keyboard := goinput.NewKeyboard()
	keyboard.Listen()
	go func() {
		for mouseMove := range mouse.OnMove {
			fmt.Println("mouse move", mouseMove.X, mouseMove.Y)
		}
	}()
	go func() {
		for mouseClick := range mouse.OnClick {
			fmt.Println("mouse click", mouseClick.Button.Name, mouseClick.Press)
		}
	}()
	go func() {
		for mouseScroll := range mouse.OnScroll {
			fmt.Println("mouse scroll", mouseScroll.Dx, mouseScroll.Dy)
		}
	}()
	go func() {
		for keyPress := range keyboard.OnPress {
			fmt.Println("keyboard press", keyPress.Name)
		}
	}()
	go func() {
		for keyRelease := range keyboard.OnRelease {
			fmt.Println("keyboard release", keyRelease.Name)
		}
	}()
	<-time.After(5 * time.Second)
	mouse.StopListen()
	keyboard.StopListen()
	fmt.Println("Stopped listener")

	fmt.Println("Testing controller...")
	for i := 0; i <= 800; i += 100 {
		mouse.Move(i, i)
		<-time.After(1 * time.Second)
	}

	toType := "Hello World!"
	for i := 0; i < len(toType); i++ {
		str := string(toType[i])
		if str == " " {
			str = "space"
		} else if str == "!" {
			str = "shift+1"
		}
		keyboard.Downup(str)
	}
	keyboard.Downup("Return")
	fmt.Println("Finished")
}
