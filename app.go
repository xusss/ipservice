package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	log "github.com/segmentio/go-log"
	"github.com/teambition/confl"
	"github.com/teambition/gear"
	"github.com/teambition/gear/logging"
	"github.com/teambition/gear/middleware/favicon"
	"github.com/wangtuanjie/ip17mon"
)

var (
	portReg = regexp.MustCompile(`^\d+$`)
)

type result struct {
	IP      string
	Status  int
	Message string
	Data    interface{}
}

func jsonAPI(ctx *gear.Context) error {
	var ip net.IP
	var res result

	callback := ctx.Query("callback")
	ipStr := ctx.Param("ip")
	if ipStr == "" {
		ip = ctx.IP()
	} else {
		ip = net.ParseIP(ipStr)
	}

	if ip == nil {
		res = result{IP: "", Status: http.StatusBadRequest, Message: "Invalid IP format"}
	} else {
		loc, err := ip17mon.Find(ip.String())
		if err != nil {
			res = result{IP: ip.String(), Status: http.StatusNotFound, Message: err.Error()}
		} else {
			res = result{IP: ip.String(), Status: http.StatusOK, Data: loc}
		}
	}

	if callback == "" {
		return ctx.JSON(res.Status, res)
	}
	return ctx.JSONP(res.Status, callback, res)

}

type config struct {
	DataPath string `json:"data_path"`
	Port     string `json:"port"`
}

func app(port, dataPath string) *gear.ServerListener {
	// init IP db
	err := ip17mon.Init(dataPath)
	if err != nil {
		panic(err)
	}

	// create app
	app := gear.New()

	// add favicon middleware
	app.Use(favicon.NewWithIco(faviconData))

	// add logger middleware
	logger := logging.New(os.Stdout)
	logger.SetLogConsume(func(log logging.Log, _ *gear.Context) {
		now := time.Now()
		delete(log, "Start")
		delete(log, "Type")
		switch res, err := json.Marshal(log); err == nil {
		case true:
			logger.Output(now, logging.InfoLevel, string(res))
		default:
			logger.Output(now, logging.WarningLevel, err.Error())
		}
	})
	app.UseHandler(logger)

	// add router middleware
	router := gear.NewRouter()
	router.Get("/json/:ip", jsonAPI)
	router.Otherwise(func(ctx *gear.Context) error {
		log := logging.FromCtx(ctx)
		log.Reset() // Reset log, don't logging for non-api request.
		return ctx.HTML(200, indexHTML)
	})
	app.UseHandler(router)

	// start app
	logging.Info("IP Service start " + port)
	return app.Start(port)
}

func checkPort(port string) string {
	if portReg.MatchString(port) {
		return ":" + port
	}
	return port
}

func main() {
	var srv *gear.ServerListener
	c := &config{}

	watcher, err := confl.NewFromEnv(c, nil)
	if err != nil {
		panic(err)
	}

	if c.Port == "" || c.DataPath == "" {
		os.Exit(1)
	}

	watcher.AddHook(func(c interface{}) {
		if cfg, ok := c.(*config); ok {
			if cfg.Port == "" || cfg.DataPath == "" {
				return
			}
			if srv != nil {
				srv.Close()
			}
			srv = app(checkPort(cfg.Port), cfg.DataPath)
		}
	})

	go watcher.GoWatch()

	srv = app(checkPort(c.Port), c.DataPath)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			watcher.Close()
			os.Exit(0)
		}
	}
}
