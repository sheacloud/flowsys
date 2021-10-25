package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sheacloud/flowsys/internal/output"
	"github.com/sheacloud/flowsys/internal/schema"
)

type PostFlowsInput struct {
	Flows []schema.FlowModel `json:"flows" binding:"required,dive"`
}

type PostFlowsOutput struct {
	Error          string `json:"error,omitempty"`
	ProcessedFlows int    `json:"processed_flows,omitempty"`
}

func PostFlows(kinesisConfig output.KinesisStreamConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var flowsRequest PostFlowsInput
		err := c.ShouldBind(&flowsRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, PostFlowsOutput{Error: err.Error()})
			return
		}

		parsedFlows := make([]schema.FlowModel, len(flowsRequest.Flows))
		for i, flow := range flowsRequest.Flows {
			flow.PopulateFields()
			if err != nil {
				c.JSON(http.StatusBadRequest, PostFlowsOutput{Error: err.Error()})
				return
			}
			parsedFlows[i] = flow
		}

		err = output.UploadFlows(parsedFlows, kinesisConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, PostFlowsOutput{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, PostFlowsOutput{ProcessedFlows: len(parsedFlows)})
	}
}

func AddFlowsRoutes(rg *gin.RouterGroup, kinesisConfig output.KinesisStreamConfig) {
	flows := rg.Group("/flows")

	flows.POST("/", PostFlows(kinesisConfig))
}
