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

func AddToConnectionQueue(app *App, Remote string, BoundaryName string, req *Request, res *Response, function func(*Request, *Response), c gnet.Conn) {
	app.ConnectionQueueLock.Lock()
	app.ConnectionQueue[Remote] = &Connection{
		BoundaryName: BoundaryName,
		req:          req,
		res:          res,
		function:     function,
	}
	app.ConnectionQueueLock.Unlock()
	if len(req.Headers["Expect"]) >= 1 && req.Headers["Expect"][0] == "100-continue" {
		c.Write(String2Slice("HTTP/1.1 100 Continue\r\n\r\n"))
	}
	AddFormData(app, Remote, req.Body, c)
}

func AddFormData(app *App, Remote string, buf []byte, c gnet.Conn) gnet.Action {
	app.ConnectionQueueLock.RLock()
	if app.ConnectionQueue[Remote] != nil {
		app.ConnectionQueue[Remote].Lock.Lock()
		FromData := bytes.Split(buf, String2Slice("--"+app.ConnectionQueue[Remote].BoundaryName))
		FormDataLength := len(FromData)
		if FormDataLength < 1 {
			app.ConnectionQueueLock.RUnlock()
			app.ConnectionQueue[Remote].Lock.Unlock()
			return gnet.Shutdown
		}
		if FormDataLength == 1 {
			if app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName] != nil && app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName].Value != nil {

				app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName].Value.Write(FromData[0])
			}
			app.ConnectionQueueLock.RUnlock()
			app.ConnectionQueue[Remote].Lock.Unlock()
			return gnet.None
		}
		if Bytes2String(FromData[0]) != "" {
			if app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName] != nil && app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName].Value != nil {
				app.ConnectionQueue[Remote].FormData[app.ConnectionQueue[Remote].LastFormDataName].Value.Write(FromData[0])
			}
		}
		for i := 1; i < FormDataLength; i++ {
			if strings.Trim(Bytes2String(FromData[i]), " \r\n") == "--" {
				app.ConnectionQueue[Remote].req.FormData = app.ConnectionQueue[Remote].FormData
				go fmt.Println("[" + time.Now().Format("2006-01-02 15:03:04") + "|" + app.ConnectionQueue[Remote].res.Req.Path + "] Form data has been parsed.")
				app.ConnectionQueue[Remote].function(app.ConnectionQueue[Remote].req, app.ConnectionQueue[Remote].res)
				if !app.ConnectionQueue[Remote].res.Chunked {
					c.Write(app.ConnectionQueue[Remote].res.GetRaw())
				}
				app.ConnectionQueueLock.RUnlock()
				app.ConnectionQueue[Remote].Lock.Unlock()
				return gnet.None
			}
			headerText, body := Bytes2String(bytes.Split(FromData[i], String2Slice("\r\n\r\n"))[0]), bytes.Join(bytes.Split(FromData[i], String2Slice("\r\n\r\n"))[1:], String2Slice("\r\n\r\n"))
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
			if app.ConnectionQueue[Remote].FormData == nil {
				app.ConnectionQueue[Remote].FormData = map[string]*FormData{}
			}
			app.ConnectionQueue[Remote].FormData[name] = &FormData{
				Name:    name,
				Value:   bytes.NewBuffer(body),
				Type:    Type,
				Headers: Headers,
			}
			app.ConnectionQueue[Remote].LastFormDataName = name
		}
		app.ConnectionQueueLock.RUnlock()
		app.ConnectionQueue[Remote].Lock.Unlock()
		return gnet.None
	}
	app.ConnectionQueueLock.RUnlock()
	return gnet.Shutdown
}
