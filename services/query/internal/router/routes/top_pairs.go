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
	topPairsQueryTemplate = "SELECT source_ipv4_address, destination_ipv4_address, %s(measure_value::bigint) as total FROM \"%s\".\"%s\" WHERE time > ago(%vs) AND measure_name = '%s' group by source_ipv4_address, destination_ipv4_address order by %s(measure_value::bigint) DESC LIMIT %v"
)

func getTopPairsQuery(input GetTopPairsInput, dbName, tableName string) string {
	return fmt.Sprintf(topPairsQueryTemplate, input.AggregateFunction, dbName, tableName, *input.TimeRelativeOffsetSeconds, input.MetricName, input.AggregateFunction, input.MaxResults)
}

type GetTopPairsInput struct {
	MetricName                string `form:"metric_name" binding:"required"`
	AggregateFunction         string `form:"aggregate_function" binding:"required"`
	TimeRelativeOffsetSeconds *int   `form:"time_relative_offset_seconds"`
	MaxResults                int    `form:"max_results" binding:"required"`
}

func (i *GetTopPairsInput) Validate(c *gin.Context) bool {
	// set default time offset to 30min
	if i.TimeRelativeOffsetSeconds == nil {
		i.TimeRelativeOffsetSeconds = aws.Int(1800)
	}

	return true
}

type GetTopPairsOutput struct {
	Error string            `json:"error,omitempty"`
	Nodes []GetTopPairsNode `json:"nodes"`
	Links []GetTopPairsLink `json:"links"`
}

type GetTopPairsLink struct {
	Source int     `json:"source"`
	Target int     `json:"target"`
	Value  float64 `json:"value"`
}

type GetTopPairsNode struct {
	Name string `json:"name"`
}

func GetTopPairs(timestreamClient *timestreamquery.TimestreamQuery, timestreamTableName, timestreamDatabaseName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestInput GetTopPairsInput
		err := c.ShouldBindQuery(&requestInput)
		if err != nil {
			c.JSON(http.StatusBadRequest, GetTopPairsOutput{Error: err.Error()})
			return
		}

		if !requestInput.Validate(c) {
			return
		}

		query := getTopPairsQuery(requestInput, timestreamDatabaseName, timestreamTableName)

		logrus.WithFields(logrus.Fields{
			"db_name":    timestreamDatabaseName,
			"table_name": timestreamTableName,
			"query":      query,
		}).Info("Querying timestream")

		output := GetTopPairsOutput{
			Links: []GetTopPairsLink{},
			Nodes: []GetTopPairsNode{},
		}

		nodes := map[string]bool{}
		nodeIndices := map[string]int{}

		currentNodeIndex := 0

		err = timestreamClient.QueryPagesWithContext(c, &timestreamquery.QueryInput{
			QueryString: aws.String(query),
		}, func(page *timestreamquery.QueryOutput, lastPage bool) bool {
			logrus.WithFields(logrus.Fields{
				"query_status": *page.QueryStatus,
				"rows":         len(page.Rows),
			}).Info("Parsing query response page")

			for _, row := range page.Rows {

				source := *row.Data[0].ScalarValue
				target := *row.Data[1].ScalarValue
				valueString := *row.Data[2].ScalarValue

				value, err := strconv.ParseFloat(valueString, 64)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err.Error(),
					}).Error("Error parsing returned value to float")
					continue
				}

				if !nodes[source] {
					output.Nodes = append(output.Nodes, GetTopPairsNode{Name: source})
					nodeIndices[source] = currentNodeIndex
					currentNodeIndex++
					nodes[source] = true
				}
				if !nodes[target] {
					output.Nodes = append(output.Nodes, GetTopPairsNode{Name: target})
					nodeIndices[target] = currentNodeIndex
					currentNodeIndex++
					nodes[target] = true
				}

				output.Links = append(output.Links, GetTopPairsLink{
					Source: nodeIndices[source],
					Target: nodeIndices[target],
					Value:  value,
				})
			}

			return true
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, GetTopPairsOutput{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, output)
	}
}
