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
	// "sync"

	"github.com/sirupsen/logrus"
)

var deviceNumMax int = 5

type Mouse struct {
	entries []deviceEntry
	termSign chan bool
	OnMove chan MouseMove
	OnClick chan MouseClick
	OnScroll chan MouseScroll
}

type Keyboard struct {
	entries []deviceEntry
	termSign chan bool
	OnPress chan KeyPress
	OnRelease chan KeyRelease
}

type deviceEntry struct {
	fd *os.File
	event *chan InputEvent
}

// NewMouse return a mouse pointer
func NewMouse() *Mouse {
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

// NewKeyboard return a keyboard pointer
func NewKeyboard() *Keyboard {
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

// New creates a new keylogger for a device path
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

// Listen enables mouse OnMove/OnScroll/OnClick channels
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
		go func (entry deviceEntry, mouse *Mouse) {
			events := entry.Read()
			L:
			for {
				select {
				case <- mouse.termSign:
					break L
				default:
				}
				select {
				case <- mouse.termSign:
					break L
				case e := <- events:
					press := e.KeyPress()
					if e.Code == 0 || e.Code == 1 {
						loc := mouse.Getmouselocation()
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
			fmt.Println("stop mouse")
		}(entry, mouse)
	}
}

// StopListen closes mouse channels
func (mouse *Mouse) StopListen() {
	var num int
	if num = deviceNumMax; num > len(mouse.entries) {
		num = len(mouse.entries)
	}
	for i := 0; i < num; i++ {
		mouse.termSign <- true
	}
	<- time.After(1 * time.Second)
	close(mouse.OnMove)
	close(mouse.OnScroll)
	close(mouse.OnClick)
}

// Listen enables keyboard OnRelease/OnPress channels
func (keyboard *Keyboard) Listen() {
	var num int
	keyboard.OnPress = make(chan KeyPress, 10)
	keyboard.OnRelease = make(chan KeyRelease, 10)
	if num = deviceNumMax; num > len(keyboard.entries) {
		num = len(keyboard.entries)
	}
	keyboard.termSign = make(chan bool, num)
	for _, entry := range keyboard.entries[0:num] {
		go func (entry deviceEntry, keyboard *Keyboard) {
			var press bool
			events := entry.Read()
			entry.event = &events
			L:
			for {
				select {
				case <- keyboard.termSign:
					break L
				default:
				}
				select {
				case <- keyboard.termSign:
					break L
				case e := <- events:
					if e.String() == "" {
						continue
					}
					if press = e.KeyPress(); press {
						select {
						case keyboard.OnPress <- KeyPress{Key{e.Code, e.String()}}:
						default:
						}
					} else {
						select {
						case keyboard.OnRelease <- KeyRelease{Key{e.Code, e.String()}}:
						default:
						}
					}
				}
			}
			fmt.Println("stop keyboard")
		}(entry, keyboard)
	}
}

// StopListen closes keyboard channels
func (keyboard *Keyboard) StopListen() {
	var num int
	if num = deviceNumMax; num > len(keyboard.entries) {
		num = len(keyboard.entries)
	}
	// for _, entry := range keyboard.entries[0:num] {
	// 	close(*entry.event)
	// }
	for i := 0; i < num; i++ {
		keyboard.termSign <- true
	}
	<- time.After(2 * time.Second)
	close(keyboard.OnPress)
	close(keyboard.OnRelease)
}

// Read from file descriptor
// Blocking call, returns channel
// Make sure to close channel when finish
func (dev *deviceEntry) Read() chan InputEvent {
	event := make(chan InputEvent)
	go func(event chan InputEvent) {
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
func (dev *deviceEntry) read() (*InputEvent, error) {
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

// eventFromBuffer parser bytes into InputEvent struct
func (dev *deviceEntry) eventFromBuffer(buffer []byte) (*InputEvent, error) {
	event := &InputEvent{}
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
