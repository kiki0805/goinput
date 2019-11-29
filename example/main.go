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
	<- time.After(5 * time.Second)
	mouse.StopListen()
	fmt.Println("stopped")
	<- time.After(5 * time.Second)

}
