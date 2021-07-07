package output

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/sheacloud/flowsys/schema"
	"github.com/sirupsen/logrus"
)

// TODO figure out how much data to upload to Kinesis - need to take into account how many open shards there are and how many ingestion services are running

type KinesisStream struct {
	streamName          string
	recordsBuffer       []types.PutRecordsRequestEntry
	recordBufferLock    sync.Mutex
	uploadLock          sync.Mutex
	kinesisClient       *kinesis.Client
	maxUploadDataLength int
	uploadsPerSecond    int
	uploadTicker        *time.Ticker
	uploadStopChannel   chan bool
}

func NewKinesisStream(streamName string, maxUploadDataLength, uploadsPerSecond int, kinesisClient *kinesis.Client) *KinesisStream {
	stream := &KinesisStream{
		streamName:          streamName,
		kinesisClient:       kinesisClient,
		maxUploadDataLength: maxUploadDataLength,
		uploadsPerSecond:    uploadsPerSecond,
		uploadStopChannel:   make(chan bool),
	}

	return stream
}

func getFlowPartitionKey(f *schema.Flow) string {
	return fmt.Sprintf("%s%v%s%v%v", net.IP(f.GetSourceIPv4Address()).String(), f.GetSourcePort(), net.IP(f.GetDestinationIPv4Address()).String(), f.GetDestinationPort(), f.GetProtocol())
}

func (k *KinesisStream) Start() {
	k.uploadTicker = time.NewTicker(time.Duration(int64(time.Second) / int64(k.uploadsPerSecond)))

	go func() {
		for {
			select {
			case <-k.uploadStopChannel:
				logrus.Info("Stopping Kinesis Uploader goroutine")
				return
			case <-k.uploadTicker.C:
				err := k.upload()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("Error uploading Kinesis buffer")
				}
			}
		}
	}()
}

func (k *KinesisStream) Stop() {
	logrus.Info("Stopping Kinesis")
	k.upload()
	k.uploadStopChannel <- true
}

func (k *KinesisStream) upload() error {
	k.uploadLock.Lock()
	k.recordBufferLock.Lock()
	defer k.recordBufferLock.Unlock()
	defer k.uploadLock.Unlock()

	// if record buffer is empty, just skip this upload
	if len(k.recordsBuffer) == 0 {
		return nil
	}

	recordsToUploadSize := 0
	recordsToUpload := make([]types.PutRecordsRequestEntry, 0)

	var newRecordsBuffer []types.PutRecordsRequestEntry
	for i, record := range k.recordsBuffer {
		if len(record.Data)+recordsToUploadSize > k.maxUploadDataLength {
			newRecordsBuffer = k.recordsBuffer[i:]
			break
		}

		if i == 500 {
			newRecordsBuffer = k.recordsBuffer[i:]
			break
		}

		recordsToUpload = append(recordsToUpload, record)
		recordsToUploadSize += len(record.Data)
		recordsToUploadSize += len(*record.PartitionKey)
	}
	if newRecordsBuffer == nil {
		newRecordsBuffer = make([]types.PutRecordsRequestEntry, 0)
	}

	k.recordsBuffer = newRecordsBuffer

	resp, err := k.kinesisClient.PutRecords(context.TODO(), &kinesis.PutRecordsInput{
		Records:    recordsToUpload,
		StreamName: aws.String(k.streamName),
	})

	if err != nil {
		// add the attempted upload records back to the front of the buffer
		k.recordsBuffer = append(recordsToUpload, k.recordsBuffer...)

		logrus.WithFields(logrus.Fields{
			"num_records":       len(recordsToUpload),
			"upload_size_bytes": recordsToUploadSize,
			"error":             err,
		}).Error("API Error uploading records to Kinesis")

		return err
	}

	if *resp.FailedRecordCount > 0 {
		// add the failed records back to the front of the buffer
		for i, record := range resp.Records {
			if record.ErrorCode != nil {
				k.recordsBuffer = append([]types.PutRecordsRequestEntry{recordsToUpload[i]}, k.recordsBuffer...)
			}
		}

		logrus.WithFields(logrus.Fields{
			"num_records":        len(recordsToUpload),
			"upload_size_bytes":  recordsToUploadSize,
			"num_failed_records": *resp.FailedRecordCount,
		}).Error("Partial error uploading records to Kinesis")

		return fmt.Errorf("Failed to upload %v records to Kinesis", *resp.FailedRecordCount)
	}

	logrus.WithFields(logrus.Fields{
		"num_records":       len(recordsToUpload),
		"upload_size_bytes": recordsToUploadSize,
	}).Info("Uploaded records to Kinesis")

	return nil
}

func (k *KinesisStream) AddFlow(flow *schema.Flow) error {
	protoData, err := proto.Marshal(flow)
	if err != nil {
		return err
	}

	protoLength := len(protoData)

	// add the length of the protobuf to the start of the message so we can read consecutive protobufs from a single file/wire elsewhere
	data := make([]byte, 4+protoLength)
	binary.BigEndian.PutUint32(data, uint32(protoLength))
	copy(data[4:], protoData)

	record := types.PutRecordsRequestEntry{
		Data:         data,
		PartitionKey: aws.String(getFlowPartitionKey(flow)),
	}

	k.recordBufferLock.Lock()
	k.recordsBuffer = append(k.recordsBuffer, record)
	k.recordBufferLock.Unlock()

	return nil
}
