package SunnyNet

import "C"
import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/qtgolang/SunnyNet/SunnyNet"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
)

// 这里面是用网络中间件来替换qkey实现快捷登录替换
var uini string
var keyi string
var Sunny = SunnyNet.NewSunny()

func GoProxy(uin string, key string, o bool) {

	uini = uin
	keyi = key
	Port := 2025
	if !o {
		Sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback).Close()
		Sunny.SetPort(Port).Close()
		err := DisableWindowsProxy()
		if err != nil {
			fmt.Println("关闭代理错误", err.Error())
			return
		}
		fmt.Println("StopProxy")
		return
	}
	Sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)

	Sunny.SetPort(Port).Start()

	fmt.Println("IE Proxy=", Sunny.SetIEProxy())
	err := Sunny.Error
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Run Port=", Port)
	//Sunny.OpenDrive(true)
	//Sunny.OpenDrive(false)
	//Sunny.ProcessALLName(true, false)
	//Sunny.ProcessALLName(true, true)
	//Sunny.ProcessAddName("QQ.exe")
	//阻止程序退出
	//select {}

}
func randomqkey() (string, string) {
	length := rand.Intn(5) + 6 // 0-4 + 6

	var uin strings.Builder
	for i := 0; i < length; i++ {
		digit := rand.Intn(10)
		uin.WriteByte(byte(digit + '0'))
	}

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length2 = 96
	var key strings.Builder
	key.Grow(length2)
	for i := 0; i < length2; i++ {
		key.WriteByte(charset[rand.Intn(len(charset))])
	}

	return uin.String(), key.String()
}
func HttpCallback(Conn SunnyNet.ConnHTTP) {
	switch Conn.Type() {
	case public.HttpSendRequest: //发起请求
		header := make(http.Header)
		url := Conn.URL()
		if strings.Contains(url, "pt_get_uins") {
			//fmt.Printf("Warn->PID: %d\n", Conn.PID())
			//header.Add("Accept", "text/plain")
			Conn.StopRequest(200, strings.ReplaceAll("var var_sso_uin_list=[{\"uin\":{uin},\"face_index\":0,\"gender\":0,\"nickname\":\"黄桃罐头\",\"client_type\":65793,\"uin_flag\":8388608,\"account\":{uin}}];ptui_getuins_CB(var_sso_uin_list);",
				"{uin}",
				uini))
			//fmt.Printf("FuckQkey->Uin=%s\n", uin)
			//Conn.SetResponseBody([]byte("clientkey=1234567"))
		}
		if strings.Contains(url, "pt_get_st") {
			header.Set("Content-Type", "application/json")
			//Set-Cookie: clientuin=2576853505; path=/; domain=ptlogin2.qq.com; Secure; SameSite=None
			//Set-Cookie: clientkey=6673281606fd6742145910e7eb5f1373e3ca117d03f24386081d43f835a76dc42803fe7b1c22bdd39f22e9eaf3576043; path=/; domain=ptlogin2.qq.com; Secure; SameSite=None
			header.Add("Set-Cookie", strings.ReplaceAll("clientuin={qq}; path=/; domain=ptlogin2.qq.com; Secure; SameSite=None", "{qq}", uini))
			header.Add("Set-Cookie", strings.ReplaceAll("clientkey={key};path=/; domain=ptlogin2.qq.com; Secure; SameSite=None", "{key}", keyi))
			//header.Add("Accept", "text/plain")
			Conn.StopRequest(200, strings.ReplaceAll("var var_sso_get_st_uin={uin: {uin}, keyindex: 19}; ptui_getst_CB(var_sso_get_st_uin);",
				"{uin}",
				uini),
				header)
			fmt.Printf("Proxy:[%s]%s\n", uini, keyi)
			//Conn.SetResponseBody([]byte("clientkey=1234567"))

		}
		if strings.Contains(url, "getface") {
			Conn.StopRequest(200, strings.ReplaceAll("pt.setHeader({\"{uin}\":\"http://q1.qlogo.cn/g?b=qq&nk={uin}&s=100\"})",
				"{uin}",
				uini))
		}
		return
	case public.HttpResponseOK:
		//bs := Conn.GetResponseBody()

		//log.Println("请求完成", Conn.GetResponseProto(), Conn.URL(), len(bs), Conn.GetResponseHeader())

		return
	case public.HttpRequestFail:
		if strings.Contains(Conn.URL(), "pt_get_uins") {
			//fmt.Printf("Warn->PID: %d\n", Conn.PID())
			//header.Add("Accept", "text/plain")
			Conn.StopRequest(200, strings.ReplaceAll("var var_sso_uin_list=[{\"uin\":{uin},\"face_index\":0,\"gender\":0,\"nickname\":\"黄桃罐头\",\"client_type\":65793,\"uin_flag\":8388608,\"account\":{uin}}];ptui_getuins_CB(var_sso_uin_list);",
				"{uin}",
				uini))
			//fmt.Printf("FuckQkey->Uin=%s\n", uin)
			//Conn.SetResponseBody([]byte("clientkey=1234567"))
		} //请求错误
		//fmt.Println("草泥马")
		//Conn.StopRequest(200, "Hello Word")
		//Conn.SetResponseBody([]byte("123456"))
		//fmt.Println(time.Now(), Conn.URL(), Conn.Error())
		return
	}
}

func WSCallback(Conn SunnyNet.ConnWebSocket) {
	switch Conn.Type() {
	case public.WebsocketConnectionOK: //连接成功
		//log.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
		return
	case public.WebsocketUserSend: //发送数据
		if Conn.MessageType() < 5 {
			//log.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
		}
		return
	case public.WebsocketServerSend: //收到数据
		if Conn.MessageType() < 5 {
			//log.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
		}
		return
	case public.WebsocketDisconnect: //连接关闭
		//log.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
		return
	default:
		return
	}
}
func TcpCallback(Conn SunnyNet.ConnTCP) {
	return
	switch Conn.Type() {
	case public.SunnyNetMsgTypeTCPAboutToConnect: //即将连接
		//mode := string(Conn.Body())
		//log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
		//修改目标连接地址
		//Conn.SetNewAddress("8.8.8.8:8080")
		return
	case public.SunnyNetMsgTypeTCPConnectOK: //连接成功
		//log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
		return
	case public.SunnyNetMsgTypeTCPClose: //连接关闭
		//log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
		return
	case public.SunnyNetMsgTypeTCPClientSend: //客户端发送数据
		//log.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	case public.SunnyNetMsgTypeTCPClientReceive: //客户端收到数据

		//log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
		return
	default:
		return
	}
}
func UdpCallback(Conn SunnyNet.ConnUDP) {

	switch Conn.Type() {
	case public.SunnyNetUDPTypeSend: //客户端向服务器端发送数据

		//log.Println("PID", Conn.PID(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改发送的数据
		//Conn.SetBody([]byte("Hello Word"))

		return
	case public.SunnyNetUDPTypeReceive: //服务器端向客户端发送数据
		//log.Println("PID", Conn.PID(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
		//修改响应的数据
		//Conn.SetBody([]byte("Hello Word"))
		return
	case public.SunnyNetUDPTypeClosed: //关闭会话
		//log.Println("PID", Conn.PID(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
		return
	}

}
func DisableWindowsProxy() error {
	// 设置 WinINet 代理 (浏览器)
	cmd1 := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",
		"/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f")
	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("设置 WinINet 代理失败: %v", err)
	}

	// 设置 WinHTTP 代理 (命令行工具)
	cmd2 := exec.Command("netsh", "winhttp", "reset", "proxy")
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("设置 WinHTTP 代理失败: %v", err)
	}

	return nil
}
