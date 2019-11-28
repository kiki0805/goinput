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
	// "sync"

	"github.com/sirupsen/logrus"
)

var deviceNumMax int = 3

type Mouse struct {
	entries []deviceEntry
	OnMove chan MouseMove
	OnClick chan MouseClick
	OnScroll chan MouseScroll
}

type Keyboard struct {
	entries []deviceEntry
	OnPress chan KeyPress
	OnRelease chan KeyRelease
}

type deviceEntry struct {
	fd *os.File
}

func NewMouse() *Mouse {
	mouse := new(Mouse)
	mouse.OnMove = make(chan MouseMove, 10)
	mouse.OnScroll = make(chan MouseScroll, 10)
	mouse.OnClick = make(chan MouseClick, 10)
	devicePath := findDevice("mouse", "touchpad", "trackpoint")
	for _, path := range devicePath {
		if strings.TrimSpace(path) == "" {
			continue
		}
		devEntry, _ := newDevice(path)
		mouse.entries = append(mouse.entries, *devEntry)
	}
	fmt.Println(mouse.entries)
	return mouse
}

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

func (mouse *Mouse) Read() {
	// var waitgroup sync.WaitGroup
	var num int
	if num = deviceNumMax; num > len(mouse.entries) {
		num = len(mouse.entries)
	}
	for _, entry := range mouse.entries[0:num] {
		// waitgroup.Add(1)
		go func (entry deviceEntry, mouse *Mouse) {
			for e := range entry.Read() {
				press := e.KeyPress()
				if e.Code == 0 || e.Code == 1 {
					loc := Getmouselocation()
					mouse.OnMove <- MouseMove{loc[0], loc[1]}
				} else if e.Code == 11 {
					dy := e.Value
					mouse.OnScroll <- MouseScroll{0, int(dy)}
				} else if e.Code == 272 {
					mouse.OnClick <- MouseClick{Button{mouseLeft, "LEFT"}, press}
				} else if e.Code == 273 {
					mouse.OnClick <- MouseClick{Button{mouseRight, "RIGHT"}, press}
				} else if e.Code == 274 {
					mouse.OnClick <- MouseClick{Button{mouseMiddle, "MIDDLE"}, press}
				}
			}
			// waitgroup.Done()
		}(entry, mouse)
	}
	// waitgroup.Wait()
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
