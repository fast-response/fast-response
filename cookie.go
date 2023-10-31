package fastresponse

import (
	"strconv"
	"strings"
	"time"
	"net/url"
)

type Cookie struct {
	Value       string
	Name        string
	Path        string
	maxAge      int64
	expires     string
	Domain      string
	HttpOnly    bool
	Secure      bool
	MaxAgeEnble bool
	readOnly    bool
}

func ParseCookies(req *Request) {
	cookieHeaders := req.Headers["cookie"]
	cookieHeadersLen := len(cookieHeaders)
	for i := 0; i < cookieHeadersLen; i++ {
		cookies := strings.Split(cookieHeaders[i], ";")
		cookiesLen := len(cookies)
		for e := 0; e < cookiesLen; e++ {
			cookieText := strings.Split(cookies[e], "=")
			name := strings.Trim(cookieText[0], " \n\r\t")
			value := strings.Trim(strings.Join(cookieText[1:], "="), " \n\r\t")
			cookie := &Cookie{
				Name:     name,
				Value:    value,
				readOnly: true,
			}
			req.Cookies[name] = cookie
		}
	}
}

func GenerateCookies(res *Response) {
	for _, cookie := range res.Cookies {
		cookieText := cookie.Name + "=" + url.PathEscape(cookie.Value)
		if cookie.expires != "" {
			cookieText += "; Expires=" + cookie.expires
		}
		if cookie.Domain != "" {
			cookieText += "; Domain=" + cookie.Domain
		}
		if cookie.Path != "" {
			cookieText += "; Path=" + cookie.Path
		}
		if cookie.HttpOnly {
			cookieText += "; HttpOnly"
		}
		if cookie.Secure {
			cookieText += "; Secure"
		}
		if cookie.MaxAgeEnble {
			cookieText += "; Max-Age=" + strconv.Itoa(int(cookie.maxAge))
		}
		res.SetHeader("set-cookie", cookieText)
	}
}

func (cookie *Cookie) SetExpires(t int64) {
	if cookie.readOnly {
		l, _ := time.LoadLocation("Europe/London")
		cookie.expires = time.Unix(t, 0).In(l).Format("Mon, 02 Jan 2006 15:04:05") + " GMT"
	}
}

func (cookie *Cookie) SetMaxAge(t int64) {
	if !cookie.readOnly {
		cookie.maxAge = t
		cookie.MaxAgeEnble = true
	}
}

func (res *Response) AddCookie(cookie *Cookie) {
	if cookie.Name != "" {
		if res.Cookies == nil {
			res.Cookies = map[string]*Cookie{}
		}
		res.Cookies[cookie.Name] = cookie
	}
}
