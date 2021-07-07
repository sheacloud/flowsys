package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamquery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sheacloud/flowsys/services/query/internal/router"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "net/http/pprof"
)

const (
	logToStdErrFlag = "logtostderr"
)

var (
	timestreamViper = viper.New()
	apiViper        = viper.New()
	prometheusViper = viper.New()
	logViper        = viper.New()

	logLevel  string
	logCaller bool

	timestreamClient *timestreamquery.TimestreamQuery

	timestreamDBName    string
	timestreamTableName string

	rootCmd = &cobra.Command{
		Use:  "query",
		Long: "Flow querying utility",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func initTimestreamClient() {
	sess := session.Must(session.NewSession())

	timestreamClient = timestreamquery.New(sess, &aws.Config{Region: aws.String(timestreamViper.GetString("region"))})
}

func initTimestreamOptions() {
	timestreamViper.SetEnvPrefix("timestream")
	timestreamViper.AutomaticEnv()

	timestreamViper.BindEnv("db_name")
	timestreamViper.SetDefault("db_name", "flowsys")

	timestreamViper.BindEnv("table_name")
	timestreamViper.SetDefault("table_name", "flows")

	timestreamViper.BindEnv("region")
	timestreamViper.SetDefault("region", "us-east-1")

	timestreamDBName = timestreamViper.GetString("db_name")
	timestreamTableName = timestreamViper.GetString("table_name")
}

func initApiOptions() {
	apiViper.SetEnvPrefix("API")
	apiViper.AutomaticEnv()

	apiViper.BindEnv("addr")
	apiViper.SetDefault("addr", "0.0.0.0")

	apiViper.BindEnv("port")
	apiViper.SetDefault("port", 8080)
}

func initPrometheusOptions() {
	prometheusViper.SetEnvPrefix("prometheus")
	prometheusViper.AutomaticEnv()

	prometheusViper.BindEnv("addr")
	prometheusViper.SetDefault("addr", "0.0.0.0")

	prometheusViper.BindEnv("port")
	prometheusViper.SetDefault("port", "9090")

	prometheusViper.BindEnv("path")
	prometheusViper.SetDefault("path", "/metrics")
}

func initLogOptions() {
	logViper.SetEnvPrefix("log")
	logViper.AutomaticEnv()

	logViper.BindEnv("level")
	logViper.SetDefault("level", "info")

	logViper.BindEnv("caller")
	logViper.SetDefault("caller", false)
}

func initLogging() {

	logrus.SetReportCaller(logViper.GetBool("caller"))
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	switch strings.ToLower(logViper.GetString("level")) {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		fmt.Printf("Invalid log level %s - valid options are trace, debug, info, warning, error, fatal, panic\n", logLevel)
		os.Exit(1)
	}
}

func init() {
	initTimestreamOptions()
	initApiOptions()
	initPrometheusOptions()
	initLogOptions()

	initLogging()

	initTimestreamClient()
}

func prometheusServer() {
	logrus.WithFields(logrus.Fields{
		"addr": prometheusViper.GetString("addr"),
		"port": prometheusViper.GetString("port"),
		"path": prometheusViper.GetString("path"),
	}).Info("Starting Prometheus...")

	http.Handle(prometheusViper.GetString("path"), promhttp.Handler())
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", prometheusViper.GetString("port")), nil))
}

func signalHandler(stopCh chan struct{}) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-signalCh:
			close(stopCh)
			return
		}
	}
}

func run() error {

	// initiate server
	ginRouter := router.GetRouter(timestreamClient, timestreamTableName, timestreamDBName)
	go func() {
		ginRouter.Run(fmt.Sprintf("%s:%v", apiViper.GetString("addr"), apiViper.GetInt("port")))
	}()

	stopCh := make(chan struct{})
	go signalHandler(stopCh)

	<-stopCh
	return nil
}

func main() {
	go prometheusServer()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
