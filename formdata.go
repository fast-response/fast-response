package fastresponse

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type FormData struct {
	Type    string
	Value   *bytes.Buffer
	Name    string
	Headers map[string][]string
}

func AddToConnectionQueue(Remote string, BoundaryName string, req *Request, res *Response, function func(*Request, *Response), c gnet.Conn) {
	ConnectionQueue[Remote] = &Connection{
		BoundaryName: BoundaryName,
		req:          req,
		res:          res,
		function:     function,
	}
	if len(req.Headers["expcet"]) >= 1 && req.Headers["expcet"][0] == "100-continue" {
		c.Write(String2Slice("HTTP/1.1 100 Continue\r\n\r\n"))
	}
	AddFormData(Remote, req.Body, c)
}

func AddFormData(Remote string, buf []byte, c gnet.Conn) gnet.Action {
	if ConnectionQueue[Remote] != nil {
		Formdata := bytes.Split(buf, String2Slice("--"+ConnectionQueue[Remote].BoundaryName))
		FormdataLength := len(Formdata)
		if FormdataLength < 1 {
			return gnet.Shutdown
		}
		if FormdataLength == 1 {
			if ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName] != nil && ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName].Value != nil {
				ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName].Value.Write(Formdata[0])
			}
			return gnet.None
		}
		if SliceBytes2String(Formdata[0]) != "" {
			if ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName] != nil && ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName].Value != nil {
				ConnectionQueue[Remote].FormData[ConnectionQueue[Remote].LastFormDataName].Value.Write(Formdata[0])
			}
		}
		for i := 1; i < FormdataLength; i++ {
			if strings.Trim(SliceBytes2String(Formdata[i]), " \r\n") == "--" {
				ConnectionQueue[Remote].req.FormData = ConnectionQueue[Remote].FormData
				go fmt.Println("[" + time.Now().Format("2006-01-02 15:03:04") + "|" + ConnectionQueue[Remote].res.Req.Uri + "] Form data has been parsed.")
				ConnectionQueue[Remote].function(ConnectionQueue[Remote].req, ConnectionQueue[Remote].res)
				if !ConnectionQueue[Remote].res.Chunked {
					c.Write(ConnectionQueue[Remote].res.GetRaw())
				}
				return gnet.None
			}
			headerText, body := SliceBytes2String(bytes.Split(Formdata[i], String2Slice("\r\n\r\n"))[0]), bytes.Join(bytes.Split(Formdata[i], String2Slice("\r\n\r\n"))[1:], String2Slice("\r\n\r\n"))
			Headers, headers := map[string][]string{}, strings.Split(headerText, "\r\n")
			headersLength := len(headers)
			for i := 1; i < headersLength; i++ {
				tmp := strings.Split(headers[i], ":")
				if len(tmp) < 2 {
					continue
				}
				if Headers[tmp[0]] == nil {
					Headers[tmp[0]] = []string{strings.TrimSpace(strings.Join(tmp[1:], ":"))}
				} else {
					Headers[tmp[0]] = append(Headers[tmp[0]], strings.TrimSpace(strings.Join(tmp[1:], ":")))
				}
			}
			if Headers["Content-Disposition"] != nil && len(Headers["Content-Disposition"]) != 1 {
				continue
			}
			Type := "text/plain"
			if Headers["Content-Type"] != nil && len(Headers["Content-Type"]) == 1 {
				Type = Headers["Content-Type"][0]
			}
			name := ""
			ls := strings.Split(Headers["Content-Disposition"][0], ";")
			lsLength := len(ls)
			for i := 0; i < lsLength; i++ {
				if strings.Trim(strings.Split(ls[i], "=")[0], "\r\n ") == "name" {
					name = strings.Trim(strings.Split(ls[i], "=")[1], "\r\n \"")
				}
			}
			if name == "" {
				continue
			}
			if ConnectionQueue[Remote].FormData == nil {
				ConnectionQueue[Remote].FormData = map[string]*FormData{}
			}
			ConnectionQueue[Remote].FormData[name] = &FormData{
				Name:    name,
				Value:   bytes.NewBuffer(body),
				Type:    Type,
				Headers: Headers,
			}
			ConnectionQueue[Remote].LastFormDataName = name
		}
		return gnet.None
	}
	return gnet.Shutdown
}
