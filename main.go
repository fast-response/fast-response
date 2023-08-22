package FastResponse

import (
    "fmt"
    "github.com/panjf2000/gnet/v2"
    "strconv"
)

type App struct {
    Router *Router
    Config *Config
}

type Config struct {
    Port int
    Host string
    LogLevel string
    Multicore bool
}

type httpServer struct {
    gnet.BuiltinEventEngine
    eng       gnet.Engine
    addr      string
    multicore bool
    App *App
    logLevel string
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
    return hs.App.Router.MatchRoutes(c)
}

func NewApp(c *Config) *App {
    r := &Router{Routes: map[string]func(*Request, *Response){}}
    return &App{Router: r, Config: c}
}

func Run(app *App) {
    if app.Config.Port == 0 {
        app.Config.Port = 8080
    }
    go fmt.Println("Running App on http://" + app.Config.Host + ":" + strconv.Itoa(app.Config.Port))
    hs := &httpServer{addr: fmt.Sprintf("tcp://%s:%d", app.Config.Host, app.Config.Port), multicore: app.Config.Multicore, App: app, logLevel: app.Config.LogLevel}
    gnet.Run(hs, hs.addr, gnet.WithMulticore(app.Config.Multicore))
}
