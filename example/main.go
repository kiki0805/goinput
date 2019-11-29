package main

import (
	"github.com/kiki0805/goinput"
	"fmt"
	"time"
	// "github.com/sirupsen/logrus"
	// "fmt"
)

func main() {
	mouse := goinput.NewMouse()
	mouse.Listen()
	keyboard := goinput.NewKeyboard()
	keyboard.Listen()
	go func() {
		for mouseMove := range mouse.OnMove {
			fmt.Println("mouse move", mouseMove.X, mouseMove.Y)
		}
	}()
	// go func() {
	// 	for mouseClick := range mouse.OnClick {
	// 		fmt.Println("mouse click", mouseClick.Button.Name, mouseClick.Press)
	// 	}
	// }()
	// go func() {
	// 	for mouseScroll := range mouse.OnScroll {
	// 		fmt.Println("mouse scroll", mouseScroll.Dx, mouseScroll.Dy)
	// 	}
	// }()
	go func() {
		for keyPress := range keyboard.OnPress {
			fmt.Println("keyboard", keyPress.Name)
		}
	}()
	// go func() {
	// 	for keyRelease := range keyboard.OnRelease {
	// 		fmt.Println("keyboard", keyRelease.Name)
	// 	}
	// }()
	<- time.After(5 * time.Second)
	mouse.StopListen()
	keyboard.StopListen()
	fmt.Println("stopped")
	<- time.After(5 * time.Second)

}
