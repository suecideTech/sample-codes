package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	"logrusLogging/echomiddleware"
	"logrusLogging/logrushook"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	panicStackSize int  = 4 << 10
	panicStackAll  bool = true
)

// setupLogrusForGCP OperationsLogging用にLogrusを初期化
func setupLogrusForGCP() {
	log.SetLevel(log.DebugLevel)         // Debugより高いレベルのログを出力
	log.SetFormatter(&log.JSONFormatter{ // GCPのLogEntryに合わせた定義
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "time",
			log.FieldKeyLevel: "severity",
			log.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	})
	// Hookでログ出力するためLogrus標準の出力は捨てる
	log.SetOutput(ioutil.Discard)
	// 標準出力：Info
	log.AddHook(&logrushook.HookGCPLog{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
		},
	})
	// 標準エラー出力：Error, Warn, Debug
	log.AddHook(&logrushook.HookGCPLog{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.ErrorLevel,
			log.WarnLevel,
			log.DebugLevel,
		},
	})
	// 標準エラー出力+ErrorReport：Panic, Fatal
	log.AddHook(&logrushook.HookGCPLog{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
		},
		ErrorReport: true,
	})
}

// httpRequest GCPのhttpRequestに合わせた定義
type httpRequest struct {
	Latency       string `json:"latency"`
	Protocol      string `json:"protocol"`
	RemoteIP      string `json:"remoteIp"`
	RequestMethod string `json:"requestMethod"`
	RequestSize   string `json:"requestSize"`
	RequestURL    string `json:"requestUrl"`
	ResponseSize  string `json:"responseSize"`
	ServerIP      string `json:"serverIp"`
	Status        string `json:"status"`
	UserAgent     string `json:"userAgent"`
}

// echoLogging EchoLogger用ログフォーマット(GCPのLogEntryに合わせた定義)
type echoLogging struct {
	Timestamp   string      `json:"time"`
	HTTPRequest httpRequest `json:"httpRequest"`
	Message     string      `json:"message"`
	// severityはCloud LoggingがhttpRequest.statusを参照して自動付与してくれる
}

// setupEchoLoggerForGCP OperationsLogging用にEchoLoggerを初期化
func setupEchoLoggerForGCP() string {
	timestamp := "${time_rfc3339_nano}"
	httprequest := httpRequest{}
	httprequest.Latency = "${latency_human}"
	httprequest.Protocol = "${protocol}"
	httprequest.RemoteIP = "${remote_ip} Forwarded-For(${header:X-Forwarded-For})"
	httprequest.RequestMethod = "${method}"
	httprequest.RequestSize = "${bytes_in}"
	httprequest.RequestURL = "${uri}"
	httprequest.ResponseSize = "${bytes_out}"
	httprequest.ServerIP = "${host}"
	httprequest.Status = "${status}"
	httprequest.UserAgent = "${user_agent}"
	message := "${remote_ip} Forwarded-For(${header:X-Forwarded-For}) ${method}[${host}${uri}] Status(${status})"

	jsonPayload := echoLogging{
		timestamp,
		httprequest,
		message}
	jsonBytes, _ := json.Marshal(jsonPayload)

	return string(jsonBytes) + "\n"
}

func main() {
	// Setup Logrus.
	setupLogrusForGCP()

	defer func() {
		if x := recover(); x != nil {
			stack := make([]byte, panicStackSize)
			length := runtime.Stack(stack, panicStackAll)
			log.Panicf("panic(%v): %s", x, stack[:length])
		}
	}()

	// Setup Echo.
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{ // GCP用のLoggerを使う
		Format: setupEchoLoggerForGCP(),
		Output: os.Stdout,
	}))
	e.Use(echomiddleware.GCPRecoverWithConfig(echomiddleware.RecoverConfig{ // // GCP用のRecoverを使う
		StackSize:       panicStackSize,
		DisableStackAll: !panicStackAll,
	}))

	e.GET("/", func(c echo.Context) error {
		log.Debugf("%s: Severity Debug.", c.Request().Method)
		log.Errorf("%s: Severity Error.", c.Request().Method)
		log.Infof("%s: Severity Info.", c.Request().Method)
		log.Warningf("%s: Severity Warning.", c.Request().Method)
		log.WithFields(logrus.Fields{
			"testField": "testValue",
		}).Info("Extend fields.")
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/fatal", func(c echo.Context) error {
		log.Fatalf("%s: Severity Fatal.", c.Request().Method)
		// Fatalの場合ここでos.Exit(1)
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/panic", func(c echo.Context) error {
		log.Panicf("%s: Severity Panic.", c.Request().Method)
		return c.String(http.StatusOK, "Hello, World!\n")
	})

	//nilPointerTesting()

	log.Infof("Echo Initialize Complete! ListenPort(80)")
	log.Fatalln(e.Start(":80"))
}

func nilPointerTesting() {
	var p *interface{}
	*p = 0
}
