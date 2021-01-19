package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	version = "1.2"

	adbAddress  *walk.TextEdit
	console     *walk.TextEdit
	startButton *walk.PushButton

	existingAdbPid = -1

	kernel32 = syscall.NewLazyDLL("kernel32.dll")
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
	if existingAdbPid == 0 {
		// kill adb server if it is created by this program
		if pid := findADBProcess(); pid > 0 {
			if p, _ := os.FindProcess(pid); p != nil {
				p.Kill()
			}
		}
	}
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
		if existingAdbPid == -1 {
			existingAdbPid = findADBProcess()
		}
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
	procCreateMutex := kernel32.NewProc("CreateMutexW")
	_, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("goscrcpy"))))
	return int(err.(syscall.Errno)) != 0
}

func findADBProcess() (pid int) {
	// github.com/mitchellh/go-ps
	handle, _, _ := kernel32.NewProc("CreateToolhelp32Snapshot").Call(0x00000002, 0)
	if handle < 0 {
		return
	}
	defer kernel32.NewProc("CloseHandle").Call(handle)
	var entry struct {
		Size              uint32
		CntUsage          uint32
		ProcessID         uint32
		DefaultHeapID     uintptr
		ModuleID          uint32
		CntThreads        uint32
		ParentProcessID   uint32
		PriorityClassBase int32
		Flags             uint32
		ExeFile           [260]uint16
	}
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := kernel32.NewProc("Process32FirstW").Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return
	}
	for {
		e := &entry
		end := 0
		for {
			if e.ExeFile[end] == 0 {
				break
			}
			end++
		}
		if syscall.UTF16ToString(e.ExeFile[:end]) == "adb.exe" {
			pid = int(e.ProcessID)
			return
		}
		ret, _, _ := kernel32.NewProc("Process32NextW").Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}
	return
}
