package fastresponse

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/panjf2000/gnet/v2"
)

type Connection struct {
	BoundaryName     string
	LastFormDataName string
	FormData         map[string]*FormData
	res              *Response
	req              *Request
	function         func(*Request, *Response)
	Lock             *sync.Mutex
}

type App struct {
	// Route, Reference Router Type
	Router *Router

	// Configuration Data, Reference Config Type
	Config *Config

	// ConnectionQueue
	ConnectionQueue map[string]*Connection

	// ConnectionQueue Lock
	ConnectionQueueLock sync.RWMutex
}

type Config struct {
	// Listening port, Default is 8080
	Port int

	// The host address being monitored will be monitored by default, including IPv6
	Host string

	// The log level is currently invalid, but it is currently under development
	LogLevel string

	// Multicore switch, which can increase speed for multicore devices
	Multicore bool

	// HTTP Proxy Mode
	ProxyMode bool
}

type httpServer struct {
	gnet.BuiltinEventEngine
	addr      string
	multicore bool
	App       *App
	logLevel  string
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	return hs.App.Router.MatchRoutes(hs.App, c)
}

func (hs *httpServer) OnClose(c gnet.Conn, err error) gnet.Action {
	hs.App.ConnectionQueue[c.RemoteAddr().String()] = nil
	return gnet.Close
}

func (hs *httpServer) OnShutdown(eng gnet.Engine) {
	hs.App.ConnectionQueue = map[string]*Connection{}
}

func NewApp(c *Config) *App {
	// New APP, Configuration must be given, but it can be empty. It should be noted that it does not actually read configuration data, and reading the configuration will be completed at startup
	r := &Router{Routes: map[string]func(*Request, *Response){}}
	return &App{Router: r, Config: c, ConnectionQueue: map[string]*Connection{}}
}

func Run(app *App) error {
	// Run the app, it will actually read the configuration and start listening to the port
	if app.Config.Port == 0 {
		app.Config.Port = 8080
	}
	go fmt.Println("Running App on http://" + app.Config.Host + ":" + strconv.Itoa(app.Config.Port))
	hs := &httpServer{addr: fmt.Sprintf("tcp://%s:%d", app.Config.Host, app.Config.Port), multicore: app.Config.Multicore, App: app, logLevel: app.Config.LogLevel}
	return gnet.Run(hs, hs.addr, gnet.WithMulticore(app.Config.Multicore))
}
