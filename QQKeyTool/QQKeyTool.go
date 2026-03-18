package QQKeyTool

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/bytedance/sonic"
	"golang.org/x/sys/windows"
)

const (
	Qzone = "https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin={uin}&clientkey={key}&u1=https://user.qzone.qq.com/{uin}/infocenter&source=panelstar&keyindex=19"
	QMail = "https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin={uin}&clientkey={key}&u1=https://wx.mail.qq.com/list/readtemplate?name=login_page.html&keyindex=19"
	Qun   = "https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin={uin}&clientkey={key}&u1=https://qun.qq.com&keyindex=19"
)

var a = Newhttpclient()

func Loginqkey(uin string, key string) bool {
	a.Getcookiebykey(uin, key, Qzone)
	if len(a.Cookies) < 2 {
		return false
	}
	fmt.Println("Cookies:", a.Cookies)

	return true
}
func Loginqzone(uin string, key string) {
	replacer := strings.NewReplacer(
		"{uin}", uin,
		"{key}", key,
	)
	urli := replacer.Replace(Qzone)
	OpenURL(urli)
}
func Loginmail(uin string, key string) {
	replacer := strings.NewReplacer(
		"{uin}", uin,
		"{key}", key,
	)
	urli := replacer.Replace(QMail)
	OpenURL(urli)
}
func Loginqun(uin string, key string) {
	replacer := strings.NewReplacer(
		"{uin}", uin,
		"{key}", key,
	)
	urli := replacer.Replace(Qun)
	OpenURL(urli)
}

func ChangeQQINFO(uin string, clientkey string, id string, company string) bool {

	a.Getcookiebykey(uin, clientkey, Qzone)
	_, pskey := Getskeybycookie(a.Cookies)

	formData := url.Values{
		"qzreferrer": {"https://user.qzone.qq.com/proxy/domain/qzonestyle.gtimg.cn/qzone/v6/setting/profile/profile.html?tab=base&g_iframeUser=1  "},
		"nickname":   {id},
		"emoji":      {""},
		"sex":        {"1"},
		"birthday":   {"1984-01-01"},
		"province":   {""},
		"city":       {""},
		"country":    {""},
		"marriage":   {"0"},
		"bloodtype":  {"5"},
		"hp":         {"0"},
		"hc":         {"0"},
		"hco":        {"0"},
		"career":     {""},
		"company":    {company},
		"cp":         {"0"},
		"cc":         {"0"},
		"cb":         {""},
		"cco":        {"0"},
		"lover":      {""},
		"islunar":    {"0"},
		"mb":         {"1"},
		"uin":        {uin},
		"pageindex":  {"1"},
		"nofeeds":    {"1"},
		"fupdate":    {"1"},
	}
	//qzreferrer=https://user.qzone.qq.com/proxy/domain/qzonestyle.gtimg.cn/qzone/v6/setting/profile/profile.html?tab=base&g_iframeUser=1&nickname=0x4261694C4&emoji=&sex=1&birthday=1984-01-01&province=21&city=1&country=1&marriage=0&bloodtype=5&hp=0&hc=0&hco=0&career=&company=&cp=0&cc=0&cb=&cco=0&lover=&islunar=0&mb=1&uin=836083700&pageindex=1&nofeeds=1&fupdate=1
	gurl := fmt.Sprintf("https://h5.qzone.qq.com/proxy/domain/w.qzone.qq.com/cgi-bin/user/cgi_apply_updateuserinfo_new?&g_tk=%d", GetGTKbyskey(pskey))

	fmt.Println(gurl)
	req, err := http.NewRequest("POST", gurl, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}
	return true
}

func SendSS(uin string, clientkey string, text string, geqian bool) bool {

	a.Getcookiebykey(uin, clientkey, Qzone)

	gq := "0"
	if geqian {
		gq = "1"
	}

	_, pskey := Getskeybycookie(a.Cookies)
	formData := url.Values{
		"qzreferrer":       {"https://user.qzone.qq.com/" + uin + "/infocenter"},
		"syn_tweet_verson": {"1"},
		"paramstr":         {"1"},
		"pic_template":     {""},
		"richtype":         {""},
		"richval":          {""},
		"special_url":      {""},
		"subrichtype":      {""},
		"con":              {"qm" + text},
		"feedversion":      {"1"},
		"ver":              {"1"},
		"ugc_right":        {"1"},
		"to_sign":          {gq},
		"hostuin":          {uin},
		"code_version":     {"1"},
		"format":           {"fs"},
	}

	//qzreferrer=https://user.qzone.qq.com/proxy/domain/qzonestyle.gtimg.cn/qzone/v6/setting/profile/profile.html?tab=base&g_iframeUser=1&nickname=0x4261694C4&emoji=&sex=1&birthday=1984-01-01&province=21&city=1&country=1&marriage=0&bloodtype=5&hp=0&hc=0&hco=0&career=&company=&cp=0&cc=0&cb=&cco=0&lover=&islunar=0&mb=1&uin=836083700&pageindex=1&nofeeds=1&fupdate=1
	gurl := fmt.Sprintf("https://user.qzone.qq.com/proxy/domain/taotao.qzone.qq.com/cgi-bin/emotion_cgi_publish_v6?&g_tk=%d", GetGTKbyskey(pskey))

	fmt.Println(gurl)
	req, err := http.NewRequest("POST", gurl, strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	// 添加Cookie到请求头
	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	return true
}

type GroupItem struct {
	Groupname string
	Qid       string
}

func GetGroupList(uin string, clientkey string) []GroupItem {
	//http://qun.qq.com/cgi-bin/qun_mgr/get_group_list?bkn=

	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)
	urli := fmt.Sprintf("http://qun.qq.com/cgi-bin/qun_mgr/get_group_list?bkn=%d", bkn)

	req, err := http.NewRequest("GET", urli, nil)
	if err != nil {
		log.Fatalf("Err %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("Err %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 读取响应体
	respp, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
	}
	fmt.Println(string(respp))

	type Group struct {
		GC    int64  `json:"gc"`
		GN    string `json:"gn"`
		Owner int64  `json:"owner"`
	}
	type Response struct {
		EC      int     `json:"ec"`
		ErrCode int     `json:"errcode"`
		EM      string  `json:"em"`
		Create  []Group `json:"create"`
		Manage  []Group `json:"manage"`
		Join    []Group `json:"join"`
	}

	var response Response

	err = sonic.Unmarshal(respp, &response)
	if err != nil {
		fmt.Printf("JSON解析错误: %v\n", err)
		return nil
	}

	result := []GroupItem{{Groupname: "💠创建的群💠"}}
	for _, group := range response.Create {
		result = append(result, GroupItem{
			Groupname: group.GN,
			Qid:       strconv.FormatInt(group.GC, 10),
		})
	}

	result = append(result, GroupItem{Groupname: "💠管理的群💠"})
	for _, group := range response.Manage {
		result = append(result, GroupItem{
			Groupname: group.GN,
			Qid:       strconv.FormatInt(group.GC, 10),
		})
	}

	result = append(result, GroupItem{Groupname: "💠加入的群💠"})
	for _, group := range response.Join {
		result = append(result, GroupItem{
			Groupname: group.GN,
			Qid:       strconv.FormatInt(group.GC, 10),
		})
	}
	return result
}

// http://web.qun.qq.com/cgi-bin/announce/get_t_list
//POST方式提交，需要Cookies，用于取群公告
//bkn=%bkn%&qid=%群号%&ft=23&s=-1&n=10&ni=1&i=1

func GetGG(uin string, clientkey string, qunn string) string {

	a.Getcookiebykey(uin, clientkey, Qun)

	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)

	//
	gurl := "http://web.qun.qq.com/cgi-bin/announce/get_t_list"

	body := fmt.Sprintf("bkn=%d&qid=%s&ft=23&s=-1&n=10&ni=1&i=1}",
		bkn, qunn)

	req, err := http.NewRequest("POST", gurl,
		strings.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	rebody, _ := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	//fmt.Println("cnm->", string(rebody))
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}
	// 定义结构体
	var result struct {
		Feeds []struct {
			U   int `json:"u"`
			Msg struct {
				Text string `json:"text"`
			} `json:"msg"`
		} `json:"feeds"`
	}

	// 解析JSON
	if err := sonic.Unmarshal(rebody, &result); err != nil {
		fmt.Println("解析错误:", err)
	}
	var builder strings.Builder
	// 输出结果
	for _, feed := range result.Feeds {
		//fmt.Printf("用户ID: %d, 文本内容: %s\n", feed.U, feed.Msg.Text)
		builder.WriteString("发送者:" + strconv.Itoa(feed.U) + "\n内容:" + feed.Msg.Text + "\n")

	}
	return strings.ReplaceAll(builder.String(), "&nbsp;", " ")

}

// Name          string
//
//	Memory        string
//	Uploadername    string
//	Uploaderuin     string
//	DownloadTimes string
//	Busid         string
//	id            string

type FileItem struct {
	Name          string
	Memory        string
	Uploadername  string
	Uploaderuin   string
	DownloadTimes string
	Busid         string
	Id            string
}

func GetFilelist(uin string, clientkey string, qunn string) []FileItem {

	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)

	gurl := "https://pan.qun.qq.com/cgi-bin/group_file/get_file_list"

	body := fmt.Sprintf("gc=%s&bkn=%d&start_index=0&cnt=50&filter_code=0&folder_id=/&show_onlinedoc_folder=1",
		qunn, bkn)

	req, err := http.NewRequest("POST", gurl, strings.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}
	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	rebody, _ := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	var FILE struct {
		EC       int `json:"ec"`
		FileList []struct {
			BusID         int    `json:"bus_id"`
			CreateTime    int64  `json:"create_time"`
			DeadTime      int    `json:"dead_time"`
			DownloadTimes int    `json:"download_times"`
			ID            string `json:"id"`
			LocalPath     string `json:"local_path"`
			MD5           string `json:"md5"`
			ModifyTime    int64  `json:"modify_time"`
			Name          string `json:"name"`
			OwnerName     string `json:"owner_name"`
			OwnerUIN      int    `json:"owner_uin"`
			ParentID      string `json:"parent_id"`
			SafeType      int    `json:"safe_type"`
			SHA           string `json:"sha"`
			SHA3          string `json:"sha3"`
			Size          int    `json:"size"`
			Type          int    `json:"type"`
			UploadSize    int    `json:"upload_size"`
		} `json:"file_list"`
		NextIndex  int `json:"next_index"`
		OpenFlag   int `json:"open_flag"`
		TotalCount int `json:"total_cnt"`
		UserRole   int `json:"user_role"`
	}

	fmt.Println(string(rebody))
	if err = sonic.Unmarshal(rebody, &FILE); err != nil {
		fmt.Println("JSON解析错误:", err)
	}

	var allfile []FileItem
	for _, file := range FILE.FileList {
		temp := []FileItem{{
			Name:          file.Name,
			Memory:        fmt.Sprintf("%.2f MB", float64(file.Size)/(1024*1024)),
			Uploadername:  file.OwnerName,
			Uploaderuin:   strconv.Itoa(file.OwnerUIN),
			DownloadTimes: strconv.Itoa(file.DownloadTimes),
			Busid:         strconv.Itoa(file.BusID),
			Id:            file.ID,
		},
		}
		allfile = append(allfile, temp...)
	}
	return allfile

}
func gettime() string {
	currentTime := time.Now()

	format := "Mon Jan _2 2006 15:04:05"

	// 格式化当前时间
	formattedTime := currentTime.Format(format)
	return formattedTime

}

func DelFile(uin string, clientkey string, qunn string, Busid string, Id string) {

	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)

	gurl := "http://pan.qun.qq.com/cgi-bin/group_file/delete_file"

	body := fmt.Sprintf("src=qpan&gc=%s&bkn=%d&bus_id=%s&file_id=%s&app_id=4&parent_folder_id=/&file_list={\"file_list\":[{\"gc\":%s,\"app_id\":4,\"bus_id\":%s,\"file_id\":\"%s\",\"parent_folder_id\":\"/\"}]}",
		qunn, bkn, Busid, Id, qunn, Busid, Id)

	req, err := http.NewRequest("POST", gurl, strings.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	// 添加Cookie到请求头
	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	//rebody, _ := io.ReadAll(resp.Body)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

}

func SendGG(uin string, clientkey string, qunn string, content string) bool {

	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)

	//
	gurl := "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice"
	//"qid=%s&bkn=%d&text=%s&pic=&pinned=0&type=1&settings={\"is_show_edit_card\":0,\"tip_window_type\":1,\"confirm_required\":1}"
	//bkn=12715790&qid=178880959&text=%E6%88%91%E6%93%8D&pinned=0&pic=Z7svqMlvXSgLcoQliavqKw6VQOiaiaEkibxicjicTN5IGE9T4&imgHeight=902&imgWidth=1207&fid=
	body := fmt.Sprintf("qid=%s&bkn=%d&text=%s&pic=&pinned=0&type=1&settings={\"is_show_edit_card\":0,\"tip_window_type\":1,\"confirm_required\":1}",
		qunn, bkn, content)

	req, err := http.NewRequest("POST", gurl,
		strings.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := a.Session.Do(req)
	if err != nil {
		log.Fatalf("发送请求失败: %v", err)
	}
	rebody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(rebody))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	return true

}
func GetBkn(skey string) int64 {
	hash := int64(5381)
	for _, c := range skey {
		hash += (hash << 5) + int64(c)
	}
	return hash & 2147483647
}

func Getskeybycookie(cookies []*http.Cookie) (skey string, pskey string) {

	var skeyCookie *http.Cookie
	var pskeyCookie *http.Cookie
	if len(cookies) <= 2 {
		fmt.Println("key失效")
		return
	}
	for _, cookie := range cookies {
		if cookie.Name == "skey" {
			skeyCookie = cookie
		}
		if cookie.Name == "p_skey" {
			pskeyCookie = cookie
		}
	}
	if skeyCookie == nil || pskeyCookie == nil {
		return "", ""
	}
	return skeyCookie.Value, pskeyCookie.Value

}
func (client *Httpclient) Getcookiebykey(uin string, clientkey string, url string) {
	replacer := strings.NewReplacer(
		"{uin}", uin,
		"{key}", clientkey,
	)
	url1 := replacer.Replace(url)

	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)

	}

	resp, err := client.Session.Do(req)
	if err != nil {
		fmt.Println("发送请求失败:", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
	}

	client.Cookies = client.Jar.Cookies(resp.Request.URL)

}
func GetGTKbyskey(skey string) int {
	hash := 5381
	for _, c := range skey {
		hash += (hash << 5) + int(c)
	}
	return hash & 0x7fffffff
}
func OpenURL(url string) error {
	furl := url
	if !strings.Contains(url, "https") && !strings.Contains(url, "http") {
		furl = "https://" + url
	}

	fmt.Println("Open:", furl)
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")
	openPtr, _ := windows.UTF16PtrFromString("open")
	furlPtr, _ := windows.UTF16PtrFromString(furl)
	ret, _, err := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(openPtr)),
		uintptr(unsafe.Pointer(furlPtr)),
		0,
		0,
		1,
	)
	if ret <= 32 {
		return fmt.Errorf("打开URL失败: %v", err)
	}
	return nil
}
