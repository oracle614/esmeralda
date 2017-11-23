package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/trace"
	"strconv"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"chuanyun.io/esmeralda/setting"
	"chuanyun.io/esmeralda/util"
	"chuanyun.io/quasimodo/config"
	"golang.org/x/sync/errgroup"
)

var exitChan = make(chan int)

var configFilePath = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")

type Server interface {
	Start()
	Shutdown(code int, reason string)
}

type EsmeraldaServerImpl struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
}

func NewEsmeraldaServer() Server {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &EsmeraldaServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func (this *EsmeraldaServerImpl) Start() {

	go listenToSystemSignals(this)

	setting.Initialize(*configFilePath)

	router := httprouter.New()
	router.GET(setting.Settings.Web.Prefix+"/", Index)
	router.GET(setting.Settings.Web.Prefix+"/hello/:name", Hello)
	panic(http.ListenAndServe(":"+strconv.FormatInt(setting.Settings.Web.Port, 10), router))
}

func (this *EsmeraldaServerImpl) Shutdown(code int, reason string) {
	logrus.Info(util.Message("Shutdown server started"))
	logrus.Exit(code)
}

func (this *EsmeraldaServerImpl) startHttpServer() {

	// g.httpServer = api.NewHttpServer()

	// err := g.httpServer.Start(g.context)

	// if err != nil {
	// 	g.log.Error("Fail to start server", "error", err)
	// 	g.Shutdown(1, "Startup failed")
	// 	return
	// }
}

func listenToSystemSignals(server Server) {
	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)
	code := 0

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		// Stops trace if profiling has been enabled
		trace.Stop()
		server.Shutdown(0, fmt.Sprintf("system signal: %s", sig))
	case code = <-exitChan:
		server.Shutdown(code, "startup error")
	}
}

func exporter(port int64, prefix string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html>
			<head><title>Metrics Exporter</title></head>
			<body>
				<h1>Metrics Exporter</h1>
				<p><a href="./metrics">Metrics</a></p>
			</body>
		</html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatal(http.ListenAndServe(":"+config.Config.Prometheus.Port, nil))
}