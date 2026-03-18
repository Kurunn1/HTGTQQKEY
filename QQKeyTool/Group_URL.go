package QQKeyTool

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
)

type GroupUrl struct {
	Text   string
	Sender string
}

func GetGroupUrl(uin string, clientkey string, qunn string) []GroupUrl {
	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)
	gurl := fmt.Sprintf("https://qun.qq.com/cgi-bin/groupchat_url_collect/get_url_collect?t=Fri %s GMT+0800 (China Standard Time)", gettime())
	body := fmt.Sprintf("bkn=%d&gc=%s&seq=0&n=50",
		bkn, qunn)

	client := &http.Client{}
	req, err := http.NewRequest("POST", gurl, strings.NewReader(body))
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	// 添加Cookie到请求头
	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	req.Header.Set("Referer", fmt.Sprintf("https://qinfo.clt.qq.com/qinfo_v3/setting.html?groupuin=%s", qunn))
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := client.Do(req)
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
	fmt.Println(string(rebody))

	// URL 项结构
	type URLItem struct {
		RawURL    string `json:"raw_url"`
		Title     string `json:"title"`
		Thumbnail string `json:"thumbnail"`
		Time      int64  `json:"time"`
		Seq       int    `json:"seq"`
		UIN       int64  `json:"uin"`
	}
	// 名称项结构
	type NameItem struct {
		Name string `json:"n"`
		UIN  int64  `json:"u"`
	}
	type ResponseData struct {
		EC        int        `json:"ec"`
		EM        string     `json:"em"`
		SrvCode   int        `json:"srv_code"`
		URLList   []URLItem  `json:"url_list"`
		Name      string     `json:"name"`
		Intro     string     `json:"intro"`
		GrpCreate int64      `json:"grp_create_time"`
		JoinTime  int64      `json:"join_time"`
		Seq       int        `json:"seq"`
		NAll      int        `json:"n_all"`
		NameList  []NameItem `json:"name_list"`
	}

	var data ResponseData
	if err := sonic.Unmarshal(rebody, &data); err != nil {
		fmt.Printf("JSON 解析错误: %v\n", err)
	}

	var result = []GroupUrl{}
	for _, url := range data.URLList {
		temp := GroupUrl{
			Text:   url.RawURL,
			Sender: strconv.FormatInt(url.UIN, 10),
		}
		result = append(result, temp)

	}

	return result
}
