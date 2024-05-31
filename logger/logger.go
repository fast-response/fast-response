package logger

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
	"unsafe"
)

var (
	ENTRY   = LogLevel(-1)
	ERROR   = LogLevel(0)
	WARNING = LogLevel(1)
	INFO    = LogLevel(2)
	DEBUG   = LogLevel(3)
)

type LogLevel int

type log struct {
	Level   LogLevel
	Message any
	Caller  string
}

type Logger struct {
	Level LogLevel
	conn  chan *log
	File  []*os.File
}

func NewLogger(level LogLevel, file ...*os.File) *Logger {
	if len(file) == 0 {
		file = []*os.File{os.Stdout}
	}
	logger := &Logger{Level: level, conn: make(chan *log), File: file}
	go logger.print()
	return logger
}

func (l *Logger) print() {
	for log := range l.conn {
		levelName := ""
		switch log.Level {
		case 0:
			levelName = "ERROR"
		case 1:
			levelName = "WARNING"
		case 2:
			levelName = "INFO"
		case 3:
			levelName = "DEBUG"
		}

		str := string2Bytes("[" + time.Now().Format("2006-01-02 15:03:04.000") + "|" + levelName + "|" + log.Caller + "] " + toString(log.Message) + "\n")

		if len(l.File) == 1 {
			l.File[0].Write(str)
			continue
		}
		fileListLength := len(l.File)
		for i := 0; i < fileListLength; i++ {
			l.File[i].Write(str)
		}
	}
}

const digits = "0123456789"

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

func smallItoa(i int) string {
	if i < 10 {
		return digits[i : i+1]
	}
	return smallsString[i*2 : i*2+2]
}

func (l *Logger) log(level LogLevel, message any) {
	// 实现日志记录逻辑
	if level <= l.Level {
		pc, _, line, _ := runtime.Caller(2)
		lineStr := ""
		if line < 100 {
			lineStr = smallItoa(line)
		} else {
			for line > 0 {
				switch line % 10 {
				case 0:
					lineStr += "0"
				case 1:
					lineStr += "1"
				case 2:
					lineStr += "2"
				case 3:
					lineStr += "3"
				case 4:
					lineStr += "4"
				case 5:
					lineStr += "5"
				case 6:
					lineStr += "6"

				case 7:
					lineStr += "7"
				case 8:
					lineStr += "8"
				case 9:
					lineStr += "9"
				}
				line /= 10 // 去除已输出的个位数
			}
		}
		l.conn <- &log{Level: level, Message: message, Caller: runtime.FuncForPC(pc).Name() + ":" + lineStr}
	}
}

func (l *Logger) Error(err ...any) {
	l.log(LogLevel(ERROR), err)
}

func (l *Logger) Warning(warn ...any) {
	l.log(LogLevel(WARNING), warn)
}

func (l *Logger) Info(info ...any) {
	l.log(LogLevel(INFO), info)
}

func (l *Logger) Debug(debug ...any) {
	l.log(LogLevel(DEBUG), debug)
}

func string2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func toString(src any) string {
	switch s := src.(type) {
	case nil:
		return ""
	case string:
		return s
	case byte:
		bs := []byte{s}
		return *(*string)(unsafe.Pointer(&bs))
	case int, int8, int16, int32, int64, uint, uint16, uint32, uint64:
		switch i := src.(type) {
		case int:
			return strconv.FormatInt(int64(i), 10)
		case int8:
			return strconv.FormatInt(int64(i), 10)
		case int16:
			return strconv.FormatInt(int64(i), 10)
		case int32:
			return strconv.FormatInt(int64(i), 10)
		case int64:
			return strconv.FormatInt(i, 10)
		case uint:
			return strconv.FormatUint(uint64(i), 10)
		case uint8:
			return strconv.FormatUint(uint64(i), 10)
		case uint16:
			return strconv.FormatUint(uint64(i), 10)
		case uint32:
			return strconv.FormatUint(uint64(i), 10)
		case uint64:
			return strconv.FormatUint(i, 10)
		}
	case float32, float64:
		switch f := src.(type) {
		case float32:
			return strconv.FormatFloat(float64(f), 'f', -1, 32)
		case float64:
			return strconv.FormatFloat(float64(f), 'f', -1, 64)
		}

	case bool:
		return strconv.FormatBool(s)
	case error:
		if s != nil {
			return s.Error()
		} else {
			return ""
		}
	case reflect.Value:
		src = s.Interface()
		return toString(src)
	case time.Time:
		return s.Format(time.RFC3339Nano)
	case fmt.Stringer:
		return s.String()
	case io.Reader:
		byt, e := io.ReadAll(s)
		if e != nil {
			panic(e)
		} else {
			return toString(byt)
		}
	case []byte:
		byts := s
		return *(*string)(unsafe.Pointer(&byts))
	case []any:
		str := ""
		ls := s
		for k := 0; k < len(ls); k++ {
			str += ", " + toString(ls[k])
		}
		if len(str) > 2 {
			return str[2:]
		}
		return str
	case any, *any:
		sv := reflect.ValueOf(src)
		if sv.Kind() == reflect.Ptr {
			return toString(sv.Elem().Interface())
		} else if sv.Kind() == reflect.Slice {
			return toString(src.([]byte))
		} else if sv.Kind() == reflect.Map {
			mapKeys := sv.MapKeys()
			mapKeysLength := len(mapKeys)
			tmp:= "{"
			for i:=0;i<mapKeysLength;i++ {
				key := toString(mapKeys[i])
			    tmp += key + ": " +  toString(sv.MapIndex(mapKeys[i]).Interface()) + ", "
			}
			return tmp + "}"
		} else {
			return "<Type " + sv.Type().String() + ">"
		}
	}
	return ""
}
