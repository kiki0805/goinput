package main

import (
	"github.com/kiki0805/goinput"
	"fmt"
	// "github.com/sirupsen/logrus"
	// "fmt"
)

func main() {
	mouse := goinput.NewMouse()
	mouse.Read()
	go func() {
		for mouseMove := range mouse.OnMove {
			fmt.Println("from chan mouse move", mouseMove.X, mouseMove.Y)
		}
	}()
	go func() {
		for mouseClick := range mouse.OnClick {
			fmt.Println("from chan mouse click", mouseClick.Button.Name, mouseClick.Press)
		}
	}()
	for mouseScroll := range mouse.OnScroll {
		fmt.Println("from chan mouse scroll", mouseScroll.Dx, mouseScroll.Dy)
	}

}
