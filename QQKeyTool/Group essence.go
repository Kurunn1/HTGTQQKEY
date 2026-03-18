package QQKeyTool

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
)

type EssenceItem struct {
	Text       string
	Sender     string
	Req_Msg    string
	Random_Msg string
}

func GetGroupessence(uin string, clientkey string, qunn string, createhtml bool) []EssenceItem {

	a.Getcookiebykey(uin, clientkey, Qun)

	//
	gurl := fmt.Sprintf("https://qun.qq.com/essence/indexPc?gc=%s", qunn)

	////body := fmt.Sprintf("bkn=%d&qid=%s&ft=23&s=-1&n=10&ni=1&i=1}",
	//	bkn, qunn)

	client := &http.Client{}
	req, err := http.NewRequest("GET", gurl,
		nil)
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	if createhtml {
		err := os.WriteFile(qunn+"Essence.html", rebody, 0666)
		if err != nil {
			return nil
		}
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	type MsgContent struct {
		MsgType      int    `json:"msg_type"`       // 1=文本，3=图片
		Text         string `json:"text,omitempty"` // 文本内容（仅文本消息有）
		ImageUrl     string `json:"image_url,omitempty"`
		FileName     string `json:"file_name,omitempty"`
		Url          string `json:"url,omitempty"`
		ThumbnailUrl string `json:"image_thumbnail_url,omitempty"` // 缩略图链接（可选）
	}

	type MsgItem struct {
		GroupCode  string       `json:"group_code"` // 群号
		MsgSeq     int64        `json:"msg_seq"`
		MsgRandom  int64        `json:"msg_random"` // 消息序号
		SenderNick string       `json:"sender_nick"`
		SenderUin  string       `json:"sender_uin"`
		MsgContent []MsgContent `json:"msg_content"` // 消息内容（可能是文本或图片）
	}

	type InitialState struct {
		MsgList []MsgItem `json:"msgList"`
	}

	startMarker := "window.__INITIAL_STATE__="
	endMarker := "</script>"
	jsonStr := GetBetweenStr(string(rebody), startMarker, endMarker)
	fmt.Println(jsonStr)

	var initialState InitialState
	if err := sonic.Unmarshal([]byte(jsonStr), &initialState); err != nil {
		fmt.Printf("JSON解析失败：%v\n", err)

	}
	result := []EssenceItem{{}}

	for _, msg := range initialState.MsgList {
		text := msg.MsgContent[0].Text + msg.MsgContent[0].ImageUrl + msg.MsgContent[0].FileName + msg.MsgContent[0].Url
		temp := []EssenceItem{{
			Text:       strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n", "\n"),
			Sender:     fmt.Sprintf("%s[%s]", msg.SenderNick, msg.SenderUin),
			Req_Msg:    strconv.FormatInt(msg.MsgSeq, 10),
			Random_Msg: strconv.FormatInt(msg.MsgRandom, 10),
		}}
		result = append(result, temp...)
	}
	return result

}
func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	} else {
		n = n + len(start) // 增加了else，不加的会把start带上
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}
func SendGroupessence(uin string, clientkey string, qunn string, msgseq string, msgrandom string, targetqid string) string {

	a.Getcookiebykey(uin, clientkey, Qun)
	skey, _ := Getskeybycookie(a.Cookies)
	bkn := GetBkn(skey)
	//
	gurl := fmt.Sprintf("https://qun.qq.com/cgi-bin/group_digest/share_digest?bkn=%d&bkn=%d&group_code=%s&msg_seq=%s&msg_random=%s&target_group_code=%s", bkn, bkn, qunn, msgseq, msgrandom, targetqid)

	////body := fmt.Sprintf("bkn=%d&qid=%s&ft=23&s=-1&n=10&ni=1&i=1}",
	//	bkn, qunn)

	client := &http.Client{}
	req, err := http.NewRequest("GET", gurl,
		nil)
	if err != nil {
		log.Fatalf("创建请求失败: %v", err)
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	//fmt.Println("pskey：", p_skey)

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")
	req.Header.Set("Referer", fmt.Sprintf("https://qun.qq.com/essence/indexPc?gc=%s", qunn))
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
	//fmt.Println("cnm->", string(rebody))
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	fmt.Println(string(rebody))

	return ""

}
