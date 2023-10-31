package fastresponse

import (
	"fmt"
	"strconv"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Response struct {
	Code    string
	headers map[string][]string
	body    []byte
	text    string
	version string
	Cookies map[string]*Cookie
	Req     *Request
	Chunked bool
	Conn    gnet.Conn
}

var Code = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	306: "Unused",
	307: "Temporary Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Time-out",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request-URI Too Large",
	415: "Unsupported Media Type",
	416: "Requested range not satisfiable",
	417: "Expectation Failed",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Time-out",
	505: "HTTP Version not supported",
}

func Time2HttpDate() string {
	l, _ := time.LoadLocation("Europe/London")
	return time.Now().In(l).Format("Mon, 02 Jan 2006 15:04:05") + " GMT"
}

func (res *Response) SetHeader(name string, content string) string {
	if name == "Date" || name == "Content-Length" {
		return ""
	}
	if res.headers[name] == nil {
		res.headers[name] = []string{content}
	} else {
		res.headers[name] = append(res.headers[name], content)
	}
	return ""
}

func (res *Response) SetBody(body []byte) {
	res.body = body
}

func (res *Response) SetText(text string) {
	res.text = text
	res.SetBody(String2Slice(text))
}

func (res *Response) SetContentType(contentType string) {
	res.headers["Content-Type"] = []string{contentType}
}

func (res *Response) SetCode(code int) {
	if Code[code] == "" {
		res.Code = "500 Internal Server Error"
	}
	res.Code = strconv.Itoa(code) + " " + Code[code]
}

func (res *Response) GetHeaders() map[string][]string {
	return res.headers
}

func (res *Response) GetHeader(name string) []string {
	if res.headers[name] == nil {
		return []string{}
	} else {
		return res.headers[name]
	}
}

func (res *Response) GetRaw() []byte {
	res.headers["Content-Length"] = []string{strconv.Itoa(len(res.body))}
	res.headers["Date"] = []string{Time2HttpDate()}
	GenerateCookies(res)
	headers := ""
	for key, val := range res.headers {
		for _, content := range val {
			headers += key + ": " + content + "\r\n"
		}
	}
	go fmt.Println("[" + time.Now().Format("2006-01-02 15:03:04") + "|" + res.Req.Uri + "] " + res.Code + " | Body Length: " + strconv.Itoa(len(res.body)))
	return BytesCombine2(String2Slice(res.version+" "+res.Code+"\r\n"+headers+"\r\n"), res.body)
}

func (res *Response) GetBody() []byte {
	return res.body
}

func (res *Response) String() string {
	return string(res.GetRaw())
}

func (res *Response) RemoveHeader(name string, content string) string {
	if name == "Date" || name == "Content-Length" {
		return ""
	}
	if res.headers[name] == nil {
		res.headers[name] = []string{content}
	} else {
		for i := 0; i < len(res.headers[name]); i++ {
			if res.headers[name][i] == content {
				res.headers[name] = append(res.headers[name][:i], res.headers[name][(i+1):]...)
				return ""
			}
		}
	}
	return ""
}

func (res *Response) PushBody(content []byte) {
	if res.Chunked {
		res.Conn.Write(BytesCombine2(String2Slice(strconv.FormatInt(int64(len(content)), 16)+"\r\n"), content, String2Slice("\r\n")))
	} else {
		res.headers["Date"] = []string{Time2HttpDate()}
		headers := ""
		for key, val := range res.headers {
			for _, content := range val {
				headers += key + ": " + content + "\r\n"
			}
		}
		res.Conn.Write(BytesCombine2(String2Slice(res.version+" "+res.Code+"\r\n"+headers+"Transfer-Encoding: chunked\r\n"+"\r\n"+strconv.FormatInt(int64(len(content)), 16)+"\r\n"), content, String2Slice("\r\n")))
		res.Chunked = true
	}
}

func (res *Response) PushBodyEnd() {
	if res.Chunked {
		res.Conn.Write(String2Slice("0\r\n\r\n"))
	}
}

func NewResponse(req *Request, Conn gnet.Conn) *Response {
	return &Response{Code: "200 " + Code[200], headers: map[string][]string{"Server": {"FastResponse"}}, version: req.Version, Req: req, Chunked: false, Conn: Conn}
}
