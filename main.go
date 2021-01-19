package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	version = "1.1"

	adbAddress  *walk.TextEdit
	console     *walk.TextEdit
	startButton *walk.PushButton
)

func init() {
	reader, writer := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if console == nil {
				continue
			}
			console.AppendText(scanner.Text() + "\r\n")
		}
	}()
	log.SetOutput(writer)
}

func main() {
	windowTitle := fmt.Sprintf("Android Remote Control (ver %s)", version)
	if alreadyRunning() {
		win.SetForegroundWindow(win.FindWindow(nil, syscall.StringToUTF16Ptr(windowTitle)))
		return
	}
	var md *walk.Dialog
	Dialog{
		AssignTo:  &md,
		Layout:    VBox{},
		Title:     windowTitle,
		MinSize:   Size{600, 400},
		FixedSize: true,
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					TextLabel{
						Text:          "Enter ADB address:",
						StretchFactor: 2,
						TextAlignment: AlignHNearVCenter,
					},
					TextEdit{
						AssignTo: &adbAddress,
						Font: Font{
							PointSize: 9,
						},
						Alignment:     AlignHCenterVCenter,
						StretchFactor: 7,
						CompactHeight: true,
						MaxSize: Size{
							Height: 10,
						},
						OnKeyUp: func(key walk.Key) {
							if key == walk.KeyReturn {
								start()
							}
						},
					},
					PushButton{
						AssignTo:      &startButton,
						Text:          "START",
						StretchFactor: 1,
						OnClicked: func() {
							start()
						},
					},
				},
			},
			TextEdit{
				AssignTo: &console,
				VScroll:  true,
				ReadOnly: true,
			},
		},
	}.Create(nil)
	dpi := float64(md.DPI()) / 96
	screenWidth := int(float64(win.GetSystemMetrics(win.SM_CXSCREEN)) / dpi)
	screenHeight := int(float64(win.GetSystemMetrics(win.SM_CYSCREEN)) / dpi)
	md.SetX((screenWidth - md.MinSize().Width) / 2)
	md.SetY((screenHeight - md.MinSize().Height) / 2)
	icon, _ := walk.NewIconFromResourceId(2)
	if icon != nil {
		md.SetIcon(icon)
	}
	md.Run()
}

func start() {
	adbAddress.SetReadOnly(true)
	startButton.SetEnabled(false)
	go func() {
		defer adbAddress.SetReadOnly(false)
		defer startButton.SetEnabled(true)
		funcs := []func() bool{}
		addr := strings.TrimSpace(adbAddress.Text())
		if addr != "" {
			funcs = append(funcs,
				run("cmd", "/c", `scrcpy\adb.exe`, "disconnect"),
				run("cmd", "/c", `scrcpy\adb.exe`, "connect", addr),
			)
		}
		funcs = append(funcs,
			run("cmd", "/c", `scrcpy\adb.exe`, "devices"),
			run("cmd", "/c", `scrcpy\scrcpy.exe`),
		)
		for _, f := range funcs {
			if f() != true {
				return
			}
		}
	}()
}

func run(name string, args ...string) func() bool {
	return func() (success bool) {
		cmd := exec.Command(name, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
			return
		}
		go logReader(stdout)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println(err)
			return
		}
		go logReader(stderr)
		if err := cmd.Start(); err != nil {
			log.Println(err)
			return
		}
		if err := cmd.Wait(); err != nil {
			log.Println(err)
			return
		}
		success = true
		return
	}
}

func logReader(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}
}

func alreadyRunning() bool {
	procCreateMutex := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW")
	_, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("goscrcpy"))))
	return int(err.(syscall.Errno)) != 0
}
