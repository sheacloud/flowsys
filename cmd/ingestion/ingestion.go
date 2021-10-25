package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/sheacloud/flowsys/internal/collection/httpapi"
	"github.com/sheacloud/flowsys/internal/output"

	_ "net/http/pprof"
)

var (
	apiViper        = viper.New()
	prometheusViper = viper.New()
	logViper        = viper.New()
	kinesisViper    = viper.New()

	logLevel string

	httpRouter *gin.Engine
	ginLambda  *ginadapter.GinLambdaV2
)

func initApiOptions() {
	apiViper.SetEnvPrefix("api")
	apiViper.AutomaticEnv()

	apiViper.BindEnv("addr")
	apiViper.SetDefault("addr", "0.0.0.0")

	apiViper.BindEnv("port")
	apiViper.SetDefault("port", 8080)
}

func initKinesisOptions() {
	kinesisViper.SetEnvPrefix("kinesis")
	kinesisViper.AutomaticEnv()

	kinesisViper.BindEnv("stream_name")
	kinesisViper.SetDefault("stream_name", "flowsys-flows")
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
	// disable klog logging to mute underlying go-ipfix library
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	klog.SetOutput(ioutil.Discard)

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
	initApiOptions()
	initPrometheusOptions()
	initLogOptions()
	initKinesisOptions()

	initLogging()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(kinesisViper.GetString("region")))
	if err != nil {
		panic(err)
	}

	kinesisClient := kinesis.NewFromConfig(cfg)

	kinesisConfig := output.KinesisStreamConfig{
		StreamName:    kinesisViper.GetString("stream_name"),
		KinesisClient: kinesisClient,
	}

	httpRouter, err = httpapi.GetRouter(kinesisConfig)
	if err != nil {
		panic(err)
	}

	if isRunningInLambda() {
		ginLambda = ginadapter.NewV2(httpRouter)
	} else {
		go func() {
			httpRouter.Run(fmt.Sprintf("%s:%v", apiViper.GetString("addr"), apiViper.GetUint("port")))
		}()
	}
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

	for range signalCh {
		logrus.Info("Received signal - shutting down...")
		close(stopCh)
	}
}

func isRunningInLambda() bool {
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
}

func Handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, event)
}

func runLambda() {
	lambda.Start(Handler)
}

func runNormal() {
	stopCh := make(chan struct{})
	go signalHandler(stopCh)

	<-stopCh
	logrus.Info("Stopping ingestion service")
}

func main() {
	go prometheusServer()

	if isRunningInLambda() {
		runLambda()
	} else {
		runNormal()
	}
}
