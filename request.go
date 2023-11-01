package fastresponse

import (
	"bytes"
	"strings"
)

type Request struct {
    // URI of the request
    Uri string

    // Headers contains the headers
    Headers map[string][]string

    // Body contains the body of the request
    Body []byte

    // Path contains the parsed path of the request
    Path string

    // Method contains the HTTP method of the request
    Method string

    // Raw contains the raw bytes of the request
    Raw []byte

    // Version of the HTTP protocol used in the request
    Version string

    // Address of the remote peer of the request
    Addr string

    // Params contains the parsed parameters of the request
    Param map[string]string

    // Cookies contains the parsed cookies of the request
    Cookies map[string]*Cookie
}

func (req *Request) GetHeader(name string) []string {
	if req.Headers[name] == nil {
		return []string{}
	} else {
		return req.Headers[name]
	}
}

func NewRequest(ReqText []byte) (*Request, string) {
	req := &Request{Raw: ReqText}
	resText := bytes.Split(ReqText, String2Slice("\r\n"))
	resTextLength := len(resText)
	if resTextLength == 0 {
		return req, "Unable to parse message"
	}
	var headers, body = []string{}, []byte{}
	for key := 0; key < resTextLength; key++ {
		val := resText[key]
		if len(val) == 0 {
			continue
		}
		if val[0] == String2Slice("\n")[0] || val[0] == String2Slice("\r")[0] {
			header := resText[0:key]
			headerLength := len(header)
			for e := 0; e < headerLength; e++ {
				headers = append(headers, SliceBytes2String(header[e]))
			}
			body = bytes.Join(resText[key:], String2Slice("\r\n"))
			req.Body = body[1:]
			break
		}
	}
	headersLength := len(headers)
	if headersLength == 0 {
		resText = bytes.Split(bytes.Join(resText, String2Slice("\r\n")), String2Slice("\r\n\r\n"))
		if resTextLength < 2 {
			return req, "Unable to parse message"
		}
		headers, _ = strings.Split(SliceBytes2String(resText[0]), "\r\n"), bytes.Join(resText[1:], String2Slice("\r\n"))
	}
	tmp := strings.Split(headers[0], " ")
	if len(tmp) != 3 {
		return req, "Unable to parse message"
	}
	if !ContainsInSlice([]string{"HTTP/1.0", "HTTP/0.9", "HTTP/1.1", "HTTP/1.2"}, tmp[2]) {
		return req, "Unsupported protocol"
	}
	req.Method, req.Uri, req.Version = tmp[0], tmp[1], tmp[2]
	req.Headers = map[string][]string{}
	for i := 1; i < headersLength; i++ {
		tmp := strings.Split(headers[i], ":")
		if len(tmp) < 2 {
			continue
		}
		if req.Headers[tmp[0]] == nil {
			req.Headers[tmp[0]] = []string{strings.TrimSpace(strings.Join(tmp[1:], ":"))}
		} else {
			req.Headers[tmp[0]] = append(req.Headers[tmp[0]], strings.TrimSpace(strings.Join(tmp[1:], ":")))
		}
	}
	req.Cookies = map[string]*Cookie{}
	ParseCookies(req)
	return req, ""
}
