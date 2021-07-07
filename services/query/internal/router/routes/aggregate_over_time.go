package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamquery"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	aggregateOverTimeQueryTemplate = "SELECT %s(measure_value::bigint) as value, BIN(time, %vs) as timestamp FROM \"%s\".\"%s\" WHERE time > ago(%vs) AND measure_name = '%s' group by BIN(time, %vs) order by timestamp ASC"
)

func getAggregateOverTimeQuery(input GetAggregateOverTimeInput, dbName, tableName string) string {
	return fmt.Sprintf(aggregateOverTimeQueryTemplate, input.AggregateFunction, input.BinSize, dbName, tableName, *input.TimeRelativeOffsetSeconds, input.MetricName, input.BinSize)
}

type GetAggregateOverTimeInput struct {
	BinSize                   int    `form:"bin_size" binding:"required"`
	MetricName                string `form:"metric_name" binding:"required"`
	AggregateFunction         string `form:"aggregate_function" binding:"required"`
	TimeRelativeOffsetSeconds *int   `form:"time_relative_offset_seconds"`
}

func (i *GetAggregateOverTimeInput) Validate(c *gin.Context) bool {
	//binsize must be between 1s and 3600s
	if i.BinSize < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, GetAggregateOverTimeOutput{Error: "BinSize must be >= 1"})
		return false
	}

	// set default time offset to 30min
	if i.TimeRelativeOffsetSeconds == nil {
		i.TimeRelativeOffsetSeconds = aws.Int(1800)
	}

	return true
}

type GetAggregateOverTimeOutput struct {
	Error      string    `json:"error,omitempty"`
	Timestamps []string  `json:"timestamps,omitempty"`
	Datapoints []float64 `json:"datapoints,omitempty"`
}

func GetAggregateOverTime(timestreamClient *timestreamquery.TimestreamQuery, timestreamTableName, timestreamDatabaseName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestInput GetAggregateOverTimeInput
		err := c.ShouldBindQuery(&requestInput)
		if err != nil {
			c.JSON(http.StatusBadRequest, GetAggregateOverTimeOutput{Error: err.Error()})
			return
		}

		if !requestInput.Validate(c) {
			return
		}

		query := getAggregateOverTimeQuery(requestInput, timestreamDatabaseName, timestreamTableName)

		logrus.WithFields(logrus.Fields{
			"db_name":    timestreamDatabaseName,
			"table_name": timestreamTableName,
			"query":      query,
		}).Info("Querying timestream")

		output := GetAggregateOverTimeOutput{
			Timestamps: []string{},
			Datapoints: []float64{},
		}

		err = timestreamClient.QueryPagesWithContext(c, &timestreamquery.QueryInput{
			QueryString: aws.String(query),
		}, func(page *timestreamquery.QueryOutput, lastPage bool) bool {
			logrus.WithFields(logrus.Fields{
				"query_status": *page.QueryStatus,
				"rows":         len(page.Rows),
			}).Info("Parsing query response page")

			for _, row := range page.Rows {

				datapointString := *row.Data[0].ScalarValue
				timestamp := *row.Data[1].ScalarValue

				datapoint, err := strconv.ParseFloat(datapointString, 64)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err.Error(),
					}).Error("Error parsing returned data point to float")
					continue
				}

				output.Timestamps = append(output.Timestamps, timestamp)
				output.Datapoints = append(output.Datapoints, datapoint)
			}

			return true
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, GetAggregateOverTimeOutput{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, output)
	}
}
