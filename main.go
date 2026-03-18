package main

import (
	"fmt"
	"os"
	"qkey/FyneGUI"
	"strings"
	"syscall"
	"unsafe"
)

func init() {
	FyneGUI.Loadgui()
}
func main() {
}

const (
	SPI_GETDESKWALLPAPER = 0x0073
	MAX_PATH             = 260
)

func getWallpaperPath() (string, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	systemParametersInfo := user32.NewProc("SystemParametersInfoW")

	var path [MAX_PATH]uint16
	ret, _, _ := systemParametersInfo.Call(
		SPI_GETDESKWALLPAPER,
		MAX_PATH,
		uintptr(unsafe.Pointer(&path[0])),
		0,
	)

	if ret == 0 {
		return "", fmt.Errorf("获取壁纸路径失败")
	}

	return syscall.UTF16ToString(path[:]), nil
}
func checkwby() bool { //通过壁纸检测微步云沙箱
	path, err := getWallpaperPath()
	if err != nil {
		fmt.Println("错误", err)
	}
	imagePath := path

	fileInfo, _ := os.Stat(imagePath)

	fileSize := fileInfo.Size()
	if strings.Contains(path, "Desktop.jpg") && !strings.Contains(path, "wallpaper") && fileSize == 46736 {
		return true
	}
	return false
}
