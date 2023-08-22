package FastResponse

import (
	"strconv"
	"time"
	"fmt"
)

type Response struct {
	Code string
	headers map[string][]string
	body []byte
	text string
	version string
	Req *Request
}

var Code = map[int]string { 
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
	t := time.Now()
	weekday := ""
	if (t.Weekday().String() == "Tuesday") || (t.Weekday().String() == "Thursday") { 
		weekday = t.Weekday().String()[0:4]
	} else {
		weekday = t.Weekday().String()[0:3]
	}
	return  weekday + t.Format(", 2 Jan 2006 15:04:05 GMT")
}

func (res *Response) SetHeader(name string, content string) string {
	if (name == "Date" || name == "Content-Length") {
		return ""
	}
	if (res.headers[name] == nil) {
		res.headers[name] = []string {content}
	} else {
		res.headers[name] = append(res.headers[name], content)
	}
	return ""
}

func (res *Response) SetBody(body []byte) {
	res.body = body
	res.headers["Content-Length"] = []string{strconv.Itoa(len(body))}
	res.headers["Date"] = []string{Time2HttpDate()}
}

func (res *Response) SetText(text string) {
	res.text = text
	res.SetBody(String2Slice(text))
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
	if (res.headers[name] == nil) {
		return []string{}
	} else {
		return res.headers[name]
	}
}

func (res *Response) GetRaw() []byte {
		headers := ""
		for key, val := range res.headers { 
			for _, content := range val {
				headers += key + ": " + content + "\r\n"
			}
		}
		go fmt.Println("[" + time.Now().Format("2006-01-02 15:03:04") + "|" + res.Req.Uri + "] " + res.Code + " | Body Length: " + strconv.Itoa(len(res.body)))
		return BytesCombine2(String2Slice(res.version + " " + res.Code + "\r\n" +headers + "\n"), res.body)
}

func (res *Response) GetBody() []byte {
	return res.body
}

func (res *Response) String() string {
	return string(res.GetRaw())
}

func NewResponse(req *Request) *Response {
	return &Response{Code: "200 "+ Code[200], headers: map[string][]string{}, version: req.Version, Req: req}
}