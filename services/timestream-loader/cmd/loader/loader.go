package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/sheacloud/flowsys/schema"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

var (
	timestreamViper = viper.New()
	logViper        = viper.New()

	logLevel  string
	logCaller bool

	timestreamClient *timestreamwrite.TimestreamWrite

	timestreamDBName    string
	timestreamTableName string
)

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

func initTimestreamClient() {
	sess := session.Must(session.NewSession())

	timestreamClient = timestreamwrite.New(sess, &aws.Config{Region: aws.String(timestreamViper.GetString("region"))})
}

func init() {
	initTimestreamOptions()
	initLogOptions()

	initLogging()

	initTimestreamClient()
}

func handler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	records := []*timestreamwrite.Record{}

	for _, record := range kinesisEvent.Records {
		//first 4 bytes of Data is the length of the protobuf, we can ignore that
		flow := &schema.Flow{}
		if err := proto.Unmarshal(record.Kinesis.Data[4:], flow); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error unmarshalling flow protobuf from Kinesis record")
			continue
		}

		records = append(records, flowToTimestreamRecords(flow)...)

		if len(records) >= 95 {
			uploadRecords(ctx, records)
			records = []*timestreamwrite.Record{}
		}
	}

	uploadRecords(ctx, records)

	return nil
}

func uploadRecords(ctx context.Context, records []*timestreamwrite.Record) error {
	if len(records) == 0 {
		return nil
	}

	_, err := timestreamClient.WriteRecordsWithContext(ctx, &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(timestreamDBName),
		TableName:    aws.String(timestreamTableName),
		Records:      records,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Error writing records to timestream")

		return err
	}

	logrus.Info("Wrote records to timestream")

	return nil
}

func flowToTimestreamRecords(flow *schema.Flow) []*timestreamwrite.Record {
	flowTime := aws.String(strconv.FormatUint(flow.GetFlowStartMilliseconds(), 10))

	sourceIPv4AddressDimension := &timestreamwrite.Dimension{
		Name:  aws.String("source_ipv4_address"),
		Value: aws.String(net.IP(flow.GetSourceIPv4Address()).String()),
	}

	destinationIPv4AddressDimension := &timestreamwrite.Dimension{
		Name:  aws.String("destination_ipv4_address"),
		Value: aws.String(net.IP(flow.GetDestinationIPv4Address()).String()),
	}

	sourcePortDimension := &timestreamwrite.Dimension{
		Name:  aws.String("source_port"),
		Value: aws.String(strconv.FormatUint(uint64(flow.GetSourcePort()), 10)),
	}

	destinationPortDimension := &timestreamwrite.Dimension{
		Name:  aws.String("destination_port"),
		Value: aws.String(strconv.FormatUint(uint64(flow.GetDestinationPort()), 10)),
	}

	protocolDimension := &timestreamwrite.Dimension{
		Name:  aws.String("protocol"),
		Value: aws.String(strconv.FormatUint(uint64(flow.GetProtocol()), 10)),
	}

	dimensions := []*timestreamwrite.Dimension{sourceIPv4AddressDimension, destinationIPv4AddressDimension, sourcePortDimension, destinationPortDimension, protocolDimension}

	octetRecord := &timestreamwrite.Record{
		Dimensions:       dimensions,
		MeasureName:      aws.String("flow_octet_count"),
		MeasureValue:     aws.String(strconv.FormatUint(flow.GetFlowOctetCount(), 10)),
		MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeBigint),
		Time:             flowTime,
		TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
	}

	packetRecord := &timestreamwrite.Record{
		Dimensions:       dimensions,
		MeasureName:      aws.String("flow_packet_count"),
		MeasureValue:     aws.String(strconv.FormatUint(flow.GetFlowPacketCount(), 10)),
		MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeBigint),
		Time:             flowTime,
		TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
	}

	reverseOctetRecord := &timestreamwrite.Record{
		Dimensions:       dimensions,
		MeasureName:      aws.String("reverse_flow_octet_count"),
		MeasureValue:     aws.String(strconv.FormatUint(flow.GetReverseFlowOctetCount(), 10)),
		MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeBigint),
		Time:             flowTime,
		TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
	}

	reversePacketRecord := &timestreamwrite.Record{
		Dimensions:       dimensions,
		MeasureName:      aws.String("reverse_flow_packet_count"),
		MeasureValue:     aws.String(strconv.FormatUint(flow.GetReverseFlowPacketCount(), 10)),
		MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeBigint),
		Time:             flowTime,
		TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
	}

	duration := flow.GetFlowEndMilliseconds() - flow.GetFlowStartMilliseconds()
	if flow.GetFlowStartMilliseconds() >= flow.GetFlowEndMilliseconds() {
		logrus.WithFields(logrus.Fields{
			"end":   flow.GetFlowEndMilliseconds(),
			"start": flow.GetFlowStartMilliseconds(),
		}).Warning("Flow ended before it started")

		duration = 0
	}

	durationRecord := &timestreamwrite.Record{
		Dimensions:       dimensions,
		MeasureName:      aws.String("flow_duration"),
		MeasureValue:     aws.String(strconv.FormatUint(duration, 10)),
		MeasureValueType: aws.String(timestreamwrite.MeasureValueTypeBigint),
		Time:             flowTime,
		TimeUnit:         aws.String(timestreamwrite.TimeUnitMilliseconds),
	}

	return []*timestreamwrite.Record{octetRecord, packetRecord, reverseOctetRecord, reversePacketRecord, durationRecord}
}

func main() {
	lambda.Start(handler)
}
