// Package goinput Capture mouse and keyboard events or control them on Linux.
// Highly dependent on xdotool.
/* 
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
*/
package goinput

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var deviceNumMax int = 5

// Mouse with captured event channel.
type Mouse struct {
	entries  []deviceEntry
	termSign chan bool
	OnMove   chan MouseMove
	OnClick  chan MouseClick
	OnScroll chan MouseScroll
}

// Keyboard with captured event channel.
type Keyboard struct {
	entries   []deviceEntry
	termSign  chan bool
	OnPress   chan KeyPress
	OnRelease chan KeyRelease
}

type deviceEntry struct {
	fd    *os.File
}

// NewMouse return a mouse pointer.
func NewMouse() *Mouse {
	checkCmdExist("xdotool")
	mouse := new(Mouse)
	devicePath := findDevice("mouse", "touchpad", "trackpoint")
	for _, path := range devicePath {
		if strings.TrimSpace(path) == "" {
			continue
		}
		devEntry, _ := newDevice(path)
		mouse.entries = append(mouse.entries, *devEntry)
	}
	return mouse
}

// NewKeyboard return a keyboard pointer.
func NewKeyboard() *Keyboard {
	checkCmdExist("xdotool")
	keyboard := new(Keyboard)
	devicePath := findDevice("keyboard")
	for _, path := range devicePath {
		if strings.TrimSpace(path) == "" {
			continue
		}
		devEntry, _ := newDevice(path)
		keyboard.entries = append(keyboard.entries, *devEntry)
	}
	return keyboard
}

// New creates a new keylogger for a device path.
func newDevice(devPath string) (*deviceEntry, error) {
	dev := &deviceEntry{}
	if !isRoot() {
		return nil, errors.New("Must be run as root")
	}
	fd, err := os.Open(devPath)
	dev.fd = fd
	return dev, err
}

func findDevice(deviceName ...string) []string {
	path := "/sys/class/input/event%d/device/name"
	resolved := "/dev/input/event%d"
	names := append([]string{}, deviceName...)
	devicePath := make([]string, len(names))
	for i := 0; i < 255; i++ {
		buff, err := ioutil.ReadFile(fmt.Sprintf(path, i))
		if err != nil {
			continue
		}
		for _, name := range names {
			if name == "mouse" {
				if strings.Contains(strings.ToLower(string(buff)), name) &&
					!strings.Contains(strings.ToLower(string(buff)), "keyboard") {
					devicePath = append(devicePath, fmt.Sprintf(resolved, i))
				}
				continue
			}
			if strings.Contains(strings.ToLower(string(buff)), name) {
				devicePath = append(devicePath, fmt.Sprintf(resolved, i))
			}
		}
	}
	return devicePath
}

func isRoot() bool {
	return syscall.Getuid() == 0 && syscall.Geteuid() == 0
}

// Listen enables mouse OnMove/OnScroll/OnClick channels.
func (mouse *Mouse) Listen() {
	var num int
	mouse.OnMove = make(chan MouseMove, 10)
	mouse.OnScroll = make(chan MouseScroll, 10)
	mouse.OnClick = make(chan MouseClick, 10)
	if num = deviceNumMax; num > len(mouse.entries) {
		num = len(mouse.entries)
	}
	mouse.termSign = make(chan bool, num)
	for _, entry := range mouse.entries[0:num] {
		go func(entry deviceEntry, mouse *Mouse) {
			events := entry.Read()
		L:
			for {
				select {
				case <-mouse.termSign:
					break L
				default:
				}
				select {
				case <-mouse.termSign:
					break L
				case e := <-events:
					press := e.KeyPress()
					if e.Code == 0 || e.Code == 1 {
						loc := mouse.Getlocation()
						select {
						case mouse.OnMove <- MouseMove{loc[0], loc[1]}:
						default:
						}
					} else if e.Code == 11 {
						dy := e.Value
						select {
						case mouse.OnScroll <- MouseScroll{0, int(dy)}:
						default:
						}
					} else if e.Code == 272 {
						select {
						case mouse.OnClick <- MouseClick{Button{mouseLeft, "LEFT"}, press}:
						default:
						}
					} else if e.Code == 273 {
						select {
						case mouse.OnClick <- MouseClick{Button{mouseRight, "RIGHT"}, press}:
						default:
						}
					} else if e.Code == 274 {
						select {
						case mouse.OnClick <- MouseClick{Button{mouseMiddle, "MIDDLE"}, press}:
						default:
						}
					}
				}
			}
		}(entry, mouse)
	}
}

// StopListen closes mouse channels.
func (mouse *Mouse) StopListen() {
	var num int
	if num = deviceNumMax; num > len(mouse.entries) {
		num = len(mouse.entries)
	}
	for i := 0; i < num; i++ {
		mouse.termSign <- true
	}
	<-time.After(time.Second)
	close(mouse.OnMove)
	close(mouse.OnScroll)
	close(mouse.OnClick)
}

// Listen enables keyboard OnRelease/OnPress channels.
func (keyboard *Keyboard) Listen() {
	var num int
	keyboard.OnPress = make(chan KeyPress, 10)
	keyboard.OnRelease = make(chan KeyRelease, 10)
	if num = deviceNumMax; num > len(keyboard.entries) {
		num = len(keyboard.entries)
	}
	keyboard.termSign = make(chan bool, num)
	for _, entry := range keyboard.entries[0:num] {
		go func(entry deviceEntry, keyboard *Keyboard) {
			events := entry.Read()
		L:
			for {
				select {
				case <-keyboard.termSign:
					break L
				default:
				}
				select {
				case <-keyboard.termSign:
					break L
				case e := <-events:
					if e.String() == "" || e.Type != evKey {
						continue
					}
					switch e.KeyPress() {
					case true:
						keyboard.OnPress <- KeyPress{Key{e.Code, e.String()}};
					case false:
						keyboard.OnRelease <- KeyRelease{Key{e.Code, e.String()}};
					}
				}
			}
		}(entry, keyboard)
	}
}

// StopListen closes keyboard channels.
func (keyboard *Keyboard) StopListen() {
	var num int
	if num = deviceNumMax; num > len(keyboard.entries) {
		num = len(keyboard.entries)
	}
	for i := 0; i < num; i++ {
		keyboard.termSign <- true
	}
	<-time.After(time.Second)
	close(keyboard.OnPress)
	close(keyboard.OnRelease)
}

// Read from file descriptor
// Blocking call, returns channel
// Make sure to close channel when finish
func (dev *deviceEntry) Read() chan inputEvent {
	event := make(chan inputEvent)
	go func(event chan inputEvent) {
		for {
			e, err := dev.read()
			if err != nil {
				logrus.Error(err)
				close(event)
				break
			}

			if e != nil {
				event <- *e
			}
		}
	}(event)
	return event
}

// read from file description and parse binary into go struct
func (dev *deviceEntry) read() (*inputEvent, error) {
	buffer := make([]byte, eventsize)
	n, err := dev.fd.Read(buffer)
	if err != nil {
		return nil, err
	}
	// no input, dont send error
	if n <= 0 {
		return nil, nil
	}
	return dev.eventFromBuffer(buffer)
}

// eventFromBuffer parser bytes into inputEvent struct
func (dev *deviceEntry) eventFromBuffer(buffer []byte) (*inputEvent, error) {
	event := &inputEvent{}
	err := binary.Read(bytes.NewBuffer(buffer), binary.LittleEndian, event)
	return event, err
}

// Close file descriptor
func (dev *deviceEntry) Close() error {
	if dev.fd == nil {
		return nil
	}
	return dev.fd.Close()
}
