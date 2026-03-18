package QQKeyTool

import (
	"net/http"
	"net/http/cookiejar"
)

type Httpclient struct {
	Session *http.Client
	Cookies []*http.Cookie
	Jar     *cookiejar.Jar
}

func Newhttpclient() *Httpclient {
	c := &http.Client{}
	jar, _ := cookiejar.New(nil)
	//if allowredirect {
	//	c.Jar = jar
	//	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
	//		return nil
	//	}
	//}
	c.Jar = jar
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}
	return &Httpclient{
		Session: c,
		Cookies: make([]*http.Cookie, 0),
		Jar:     jar,
	}
}
