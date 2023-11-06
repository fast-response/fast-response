package fastresponse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Router struct {
	Routes map[string]func(*Request, *Response)
}

type Connection struct {
	BoundaryName     string
	LastFormDataName string
	FormData         map[string]*FormData
	res              *Response
	req              *Request
	function         func(*Request, *Response)
}

var ConnectionQueue = map[string]*Connection{}

func (r *Router) Add(rule string, function func(*Request, *Response)) {
	r.Routes[rule] = function
}

func (r *Router) To(uri string) func(*Request, *Response) {
	return func(req *Request, res *Response) {
		res.SetHeader("Location", uri)
		res.SetCode(301)
	}
}

func (r *Router) MatchRoutes(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	ok := AddFormData(c.RemoteAddr().String(), buf, c)
	if ok == gnet.None {
		return gnet.None
	}
	req, err := NewRequest(buf)
	if err != "" {
		go fmt.Println("[" + time.Now().Format("2006-01-02 15:03:04") + "] " + err)
		return gnet.Close
	}
	res := NewResponse(req, c)
	for rule, function := range r.Routes {
		match, dict := r.MatchRoute(req.Uri, rule)
		if match {
			req.Param = dict
			if !req.AddToConnectionQueue(c.RemoteAddr().String(), res, function, c) {
				function(req, res)
				if !res.Chunked {
					c.Write(res.GetRaw())
				}
			}
			return gnet.None
		}
	}
	GetErrPage(req, res, Code[404], 404)
	c.Write(res.GetRaw())
	return gnet.None
}

func (r *Router) MatchRoute(uri string, rule string) (bool, map[string]string) {
	uriList := strings.Split(uri, "/")
	ruleList := strings.Split(rule, "/")
	result := map[string]string{}
	e := 0
	if len(ruleList) == 0 {
		uriList = []string{"", ""}
	}
	ruleLength := len(ruleList)
	if len(uriList) == 0 {
		uriList = []string{"", ""}
	}
	for b := 0; b < ruleLength; b++ {
		i := ruleList[b]
		if len(uriList) <= e {
			return false, result
		}
		if i == "" {
			if uriList[e] != i {
				return false, result
			}
			e += 1
			continue
		}
		if SliceByte2String(i[0]) == "{" && SliceByte2String(i[len(i)-1]) == "}" {
			if strings.ContainsRune(i[1:len(i)-1], '*') {
				paramName := strings.Split(i[1:len(i)-1], "*")
				num, err := strconv.Atoi(paramName[1])
				if err == nil {
					if len(uriList)-1 < (num + e - 1) {
						return false, result
					}
					result[paramName[0]] = strings.Join(uriList[e:(num+e)], "/")
					e += num
				} else {
					if b == len(ruleList)-1 {
						if (len(uriList)-e) <= 0 || (uriList[e] == "") {
							return false, result
						}
						result[paramName[0]] = strings.Join(uriList[e:], "/")
						e += (len(uriList) - e)
					} else {
						if paramName[1] == "" {
							if SliceByte2String(ruleList[b+1][0]) != "{" && SliceByte2String(ruleList[b+1][len(ruleList[b+1])-1]) != "}" {
								n := e
								for {
									if (len(uriList) - n) <= 0 {
										return false, result
									}
									if uriList[n] == ruleList[b+1] {
										break
									}
									n += 1
								}
								result[paramName[0]] = strings.Join(uriList[e:n], "/")
								e = n
							} else {
								return false, result
							}
						} else {
							return false, result
						}
					}
				}
			} else {
				if (len(uriList)-e) > 0 && (uriList[e] != "") {
					result[i[1:len(i)-1]] = uriList[e]
					e += 1
				} else {
					return false, result
				}
			}
		} else {
			if uriList[e] != i {
				return false, result
			}
			e += 1
		}
	}
	if len(uriList)-1 >= e {
		return false, result
	}
	return true, result
}
