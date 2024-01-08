package fastresponse

var errPages = map[int]func(*Request, *Response, string, int){}

func GetErrPage(req *Request, res *Response, err string, errCode int) {
	if errPages[errCode] != nil {
		errPages[errCode](req, res, err, errCode)
	} else {
		defaultErrPage(req, res, err, errCode)
	}
}

func SetErrPage(errCode int, function func(*Request, *Response, string, int)) {
	errPages[errCode] = function
}

func defaultErrPage(req *Request, res *Response, err string, errCode int) {
	res.SetText(err)
	res.SetCode(errCode)
}
