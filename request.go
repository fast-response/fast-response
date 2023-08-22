package FastResponse

import (
	"bytes"
	"strings"
)

type Request struct {
	Uri string
	Headers map[string][]string
	Body []byte
	Path string
	Method string
	Raw []byte
	Version string
	Addr string
	Param map[string]string
}

func (req *Request) GetHeader(name string) []string {
	if (req.Headers[name] == nil) {
		return []string{}
	} else {
		return req.Headers[name]
	}
}

func NewRequest(ReqText []byte) (*Request, string){
	req := &Request{Raw: ReqText}
	resText := bytes.Split(ReqText,  String2Slice("\r\n"))
	if len(resText) == 0 {
		return req, "Unable to parse message"
	}
	var headers, body = []string{}, []byte{}
	for key := 0;key < len(resText);key++ {
		val := resText[key]
		if len(val) == 0 {
			continue
		}
		if val[0] == byte('\n') || val[0] == byte('\r') {
			headers, body = strings.Split(SliceBytes2String(bytes.Join(resText[0 : key], String2Slice("\r\n"))), "\r\n"), bytes.Join(resText[key:],  String2Slice("\r\n"))
			req.Body = body[1:]
			break
		}
	}
	if len(headers) == 0 {
		resText = bytes.Split(bytes.Join(resText, String2Slice("\r\n")), String2Slice("\r\n\r\n"))
		if len(resText) < 2 {
			return req, "Unable to parse message"
		}
		headers, body = strings.Split(SliceBytes2String(resText[0]), "\r\n"), bytes.Join(resText[1:],  String2Slice("\r\n"))
	}
	tmp := strings.Split(headers[0], " ")
	if len(tmp) != 3 {
		return req, "Unable to parse message"
	}
	if ContainsInSlice([]string {"HTTP/1.0", "HTTP/0.9", "HTTP/1.1", "HTTP/1.2"}, tmp[2]) == false {
		return req, "Unsupported protocol"
	}
	req.Method, req.Uri, req.Version = tmp[0], tmp[1], tmp[2]
	req.Headers = map[string][]string{}
	for i  := 1;i < len(headers);i++ {
		tmp := strings.Split(headers[i], ":")
		if len(tmp) < 2 {
			continue
		}
		if (req.Headers[tmp[0]] == nil) {
			req.Headers[tmp[0]] = []string {strings.TrimSpace(strings.Join(tmp[1:], ":"))}
		} else {
			req.Headers[tmp[0]] = append(req.Headers[tmp[0]], strings.TrimSpace(strings.Join(tmp[1:], ":")))
		}
	}
	return req, ""
}