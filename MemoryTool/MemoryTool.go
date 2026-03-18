package MemoryTool

import (
	"fmt"
	"qkey/FyneGUI"
	"strings"
	"syscall"
	"unsafe"
)

const (
	MEM_COMMIT                = 0x00001000
	PAGE_READONLY             = 0x02
	PAGE_READWRITE            = 0x04
	PAGE_EXECUTE_READ         = 0x20
	PAGE_EXECUTE_READWRITE    = 0x40
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	PROCESS_VM_OPERATION      = 0x0008

	ERROR_SUCCESS           = 0
	SE_PRIVILEGE_ENABLED    = 0x00000002
	SE_DEBUG_NAME           = "SeDebugPrivilege"
	TOKEN_ADJUST_PRIVILEGES = 0x0020
	TOKEN_QUERY             = 0x0008
)

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type_             uint32
}

type LUID struct {
	LowPart  uint32
	HighPart int32
}

type LUID_AND_ATTRIBUTES struct {
	luid       LUID
	attributes uint32
}

type TOKEN_PRIVILEGES struct {
	privilege_count uint32
	privileges      [1]LUID_AND_ATTRIBUTES
}

func HexToStr(hexStr string) string {
	result := make([]byte, 0, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		result = append(result, b)
	}
	return string(result)
}

var (
	kernel32 = syscall.NewLazyDLL(FyneGUI.XorDecrypt([]byte{0x09, 0x04, 0x1B, 0x02, 0x54, 0x0E, 0x52, 0x5B, 0x42, 0x56, 0x0E, 0x0D}))
	advapi32 = syscall.NewLazyDLL(FyneGUI.XorDecrypt([]byte{0x03, 0x05, 0x1F, 0x0D, 0x41, 0x0B, 0x52, 0x5B, 0x42, 0x56, 0x0E, 0x0D}))

	OpenProcess       = kernel32.NewProc(FyneGUI.XorDecrypt([]byte{0x2D, 0x11, 0x0C, 0x02, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11}))
	CloseHandle       = kernel32.NewProc(FyneGUI.XorDecrypt([]byte{0x21, 0x0D, 0x06, 0x1F, 0x54, 0x2A, 0x00, 0x07, 0x08, 0x5E, 0x07}))
	ReadProcessMemory = kernel32.NewProc(FyneGUI.XorDecrypt([]byte{0x30, 0x04, 0x08, 0x08, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11, 0x2C, 0x0C, 0x01, 0x5C, 0x10, 0x18}))
	VirtualQueryEx    = kernel32.NewProc(FyneGUI.XorDecrypt([]byte{0x34, 0x08, 0x1B, 0x18, 0x44, 0x03, 0x0D, 0x38, 0x19, 0x57, 0x10, 0x18, 0x2C, 0x14}))
	GetLastError      = kernel32.NewProc(FyneGUI.XorDecrypt([]byte{0x25, 0x04, 0x1D, 0x20, 0x50, 0x11, 0x15, 0x2C, 0x1E, 0x40, 0x0D, 0x13}))

	OpenProcessToken      = advapi32.NewProc(FyneGUI.XorDecrypt([]byte{0x2D, 0x11, 0x0C, 0x02, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11, 0x35, 0x06, 0x07, 0x56, 0x0C}))
	LookupPrivilegeValue  = advapi32.NewProc(FyneGUI.XorDecrypt([]byte{0x2E, 0x0E, 0x06, 0x07, 0x44, 0x12, 0x31, 0x1B, 0x05, 0x44, 0x0B, 0x0D, 0x0C, 0x0B, 0x56, 0x34, 0x00, 0x05, 0x19, 0x51, 0x35}))
	AdjustTokenPrivileges = advapi32.NewProc(FyneGUI.XorDecrypt([]byte{0x23, 0x05, 0x03, 0x19, 0x42, 0x16, 0x35, 0x06, 0x07, 0x57, 0x0C, 0x31, 0x1B, 0x05, 0x45, 0x0B, 0x0D, 0x0C, 0x0B, 0x51, 0x11}))
)

func enableDebugPrivilege() error {
	var tokenHandle syscall.Handle

	currentProcess, err := syscall.GetCurrentProcess()
	if err != nil {
		return fmt.Errorf("获取当前进程句柄失败: %v", err)
	}

	r1, _, err := OpenProcessToken.Call(
		uintptr(currentProcess),
		uintptr(TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY),
		uintptr(unsafe.Pointer(&tokenHandle)),
	)
	if r1 == 0 {
		return fmt.Errorf("OpenProcessToken 失败: %v", err)
	}
	defer CloseHandle.Call(uintptr(tokenHandle))

	var luid LUID
	seDebugName, _ := syscall.UTF16PtrFromString(SE_DEBUG_NAME)
	r1, _, err = LookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(seDebugName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if r1 == 0 {
		return fmt.Errorf("LookupPrivilegeValue 失败: %v", err)
	}

	tokenPrivileges := TOKEN_PRIVILEGES{
		privilege_count: 1,
		privileges: [1]LUID_AND_ATTRIBUTES{
			{
				luid:       luid,
				attributes: SE_PRIVILEGE_ENABLED,
			},
		},
	}

	r1, _, err = AdjustTokenPrivileges.Call(
		uintptr(tokenHandle),
		0,
		uintptr(unsafe.Pointer(&tokenPrivileges)),
		0,
		0,
		0,
	)
	if r1 == 0 {
		return fmt.Errorf("AdjustTokenPrivileges 失败: %v", err)
	}

	r1, _, _ = GetLastError.Call()
	if r1 != ERROR_SUCCESS {
		return fmt.Errorf("启用调试权限失败，错误码: %d", r1)
	}

	return nil
}

func OpenProcess_(processID uint32) (syscall.Handle, error) {
	if err := enableDebugPrivilege(); err != nil {
	}

	access := PROCESS_QUERY_INFORMATION | PROCESS_VM_READ | PROCESS_VM_OPERATION
	r1, _, err := OpenProcess.Call(
		uintptr(access),
		0,
		uintptr(processID),
	)
	if r1 == 0 {
		access = PROCESS_QUERY_INFORMATION | PROCESS_VM_READ
		r1, _, err = OpenProcess.Call(
			uintptr(access),
			0,
			uintptr(processID),
		)
		if r1 == 0 {
			return 0, err
		}
	}

	return syscall.Handle(r1), nil
}

func readMemory(handle syscall.Handle, address uintptr, size uintptr) ([]byte, error) {
	buffer := make([]byte, size)
	var bytesRead uintptr
	r1, _, err := ReadProcessMemory.Call(
		uintptr(handle),
		address,
		uintptr(unsafe.Pointer(&buffer[0])),
		size,
		uintptr(unsafe.Pointer(&bytesRead)),
	)
	if r1 == 0 {
		return nil, err
	}
	if bytesRead != size {
		return nil, fmt.Errorf("cnmm %d, %d", size, bytesRead)
	}

	return buffer, nil
}

func queryMemory(handle syscall.Handle, address uintptr) (*MEMORY_BASIC_INFORMATION, error) {
	var mbi MEMORY_BASIC_INFORMATION
	r1, _, err := VirtualQueryEx.Call(
		uintptr(handle),
		address,
		uintptr(unsafe.Pointer(&mbi)),
		unsafe.Sizeof(mbi),
	)
	if r1 == 0 {
		return nil, err
	}

	return &mbi, nil
}

func patternMatch(data, pattern []byte, mask string) bool {
	if len(data) < len(pattern) || len(mask) != len(pattern) {
		return false
	}

	for i := 0; i < len(pattern); i++ {
		if mask[i] != '?' && data[i] != pattern[i] {
			return false
		}
	}

	return true
}

func SearchPattern(handle syscall.Handle, pattern []byte, mask string, startAddr, endAddr uintptr) []uintptr {
	var results []uintptr
	currentAddr := startAddr

	for currentAddr < endAddr {
		mbi, err := queryMemory(handle, currentAddr)
		if err != nil {
			currentAddr += 0x1000
			continue
		}

		if mbi.State == MEM_COMMIT && (mbi.Protect&(PAGE_READONLY|PAGE_READWRITE|PAGE_EXECUTE_READ|PAGE_EXECUTE_READWRITE) != 0) {
			regionSize := mbi.RegionSize
			if currentAddr+regionSize > endAddr {
				regionSize = endAddr - currentAddr
			}

			memoryData, err := readMemory(handle, currentAddr, regionSize)
			if err == nil {
				for i := 0; i <= len(memoryData)-len(pattern); i++ {
					if patternMatch(memoryData[i:], pattern, mask) {
						results = append(results, currentAddr+uintptr(i))
					}
				}
			}
		}

		currentAddr += mbi.RegionSize
	}

	return results
}

func GenerateString(content string, count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(content)
	}
	return builder.String()
}
