package output

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/sheacloud/flowsys/internal/schema"
	"github.com/sirupsen/logrus"
)

var (
	MaxUploadSize       = 5 * 1024 * 1024
	MaxRecordsPerUpload = 500
)

// TODO figure out how much data to upload to Kinesis - need to take into account how many open shards there are and how many ingestion services are running

type KinesisStreamConfig struct {
	StreamName    string
	KinesisClient *kinesis.Client
}

func UploadFlows(flows []schema.FlowModel, config KinesisStreamConfig) error {
	if len(flows) == 0 {
		return nil
	}

	recordsToUpload := make([]types.PutRecordsRequestEntry, len(flows))

	for i, flow := range flows {
		recordRequestEntry, err := getFlowRecordRequetEntry(flow)
		if err != nil {
			return err
		}

		recordsToUpload[i] = *recordRequestEntry
	}

	for {
		nextBatch, remainingRecords := getNextRecordBatch(recordsToUpload)
		recordsToUpload = remainingRecords

		failedRecords, err := uploadToKinesis(nextBatch, config)
		if err != nil {
			return err
		}
		if failedRecords != nil {
			recordsToUpload = append(recordsToUpload, failedRecords...)
		}

		logrus.WithFields(logrus.Fields{
			"uploaded_records": len(nextBatch),
		}).Info("uploaded batch of records to kinesis")

		if len(recordsToUpload) == 0 {
			break
		}
	}

	return nil
}

// return next batch, remaining records
func getNextRecordBatch(records []types.PutRecordsRequestEntry) ([]types.PutRecordsRequestEntry, []types.PutRecordsRequestEntry) {

	currentBatchSize := 0
	for i, record := range records {
		if (currentBatchSize+len(record.Data)) < MaxUploadSize && i < MaxRecordsPerUpload {
			currentBatchSize += len(record.Data)
			continue
		}

		return records[:i], records[i:]
	}

	// all records are within next batch
	return records, nil
}

func uploadToKinesis(records []types.PutRecordsRequestEntry, config KinesisStreamConfig) ([]types.PutRecordsRequestEntry, error) {
	resp, err := config.KinesisClient.PutRecords(context.TODO(), &kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String(config.StreamName),
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"attempted_records": len(records),
			"error":             err,
		}).Error("API Error uploading records to Kinesis")
		return nil, err
	}

	badRecords := make([]types.PutRecordsRequestEntry, 0)
	for i, record := range resp.Records {
		if record.ErrorCode != nil {
			badRecords = append(badRecords, records[i])
		}
	}
	if len(badRecords) > 0 {
		logrus.WithFields(logrus.Fields{
			"attempted_records": len(records),
			"failed_records":    len(badRecords),
		}).Error("failed to upload all records to kinesis")
		return badRecords, nil
	}

	return nil, nil
}

func getFlowRecordRequetEntry(flow schema.FlowModel) (*types.PutRecordsRequestEntry, error) {
	data, err := flow.ToJSON()
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	hash.Write(data)

	return &types.PutRecordsRequestEntry{
		Data:         data,
		PartitionKey: aws.String(fmt.Sprintf("%x", hash.Sum(nil))),
	}, nil
}
