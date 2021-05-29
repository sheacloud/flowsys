package routes

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sheacloud/flowsys/services/ingestion/internal/enrichment"
	"github.com/sheacloud/flowsys/services/ingestion/internal/output"
	"github.com/sheacloud/flowsys/services/ingestion/internal/schema"
)

type FlowModel struct {
	SourceIP               string  `json:"source_ip" binding:"required"`
	DestinationIP          string  `json:"destination_ip" binding:"required"`
	SourcePort             *uint16 `json:"source_port" binding:"required"`
	DestinationPort        *uint16 `json:"destination_port" binding:"required"`
	Protocol               *uint8  `json:"protocol" binding:"required"`
	FlowStartMilliseconds  *uint64 `json:"flow_start_milliseconds" binding:"required"`
	FlowEndMilliseconds    *uint64 `json:"flow_end_milliseconds" binding:"required"`
	FlowOctetCount         *uint64 `json:"flow_octet_count" binding:"required"`
	FlowPacketCount        *uint64 `json:"flow_packet_count" binding:"required"`
	ReverseFlowOctetCount  *uint64 `json:"reverse_flow_octet_count"`
	ReverseFlowPacketCount *uint64 `json:"reverse_flow_packet_count"`
}

func (f *FlowModel) ToFlow() (*schema.Flow, error) {
	flow := &schema.Flow{}

	sourceIP := net.ParseIP(f.SourceIP)
	if sourceIP == nil {
		return nil, fmt.Errorf("SourceIP is an invalid IP address: %s", f.SourceIP)
	}
	flow.SourceIPv4Address = []byte(sourceIP)

	destinationIP := net.ParseIP(f.DestinationIP)
	if destinationIP == nil {
		return nil, fmt.Errorf("DestinationIP is an invalid IP address: %s", f.DestinationIP)
	}
	flow.DestinationIPv4Address = []byte(destinationIP)

	flow.SourcePort = uint32(*f.SourcePort)
	flow.DestinationPort = uint32(*f.DestinationPort)
	flow.Protocol = uint32(*f.Protocol)
	flow.FlowStartMilliseconds = *f.FlowStartMilliseconds
	flow.FlowEndMilliseconds = *f.FlowEndMilliseconds
	flow.FlowOctetCount = *f.FlowOctetCount
	flow.FlowPacketCount = *f.FlowPacketCount
	flow.ReverseFlowOctetCount = *f.ReverseFlowOctetCount
	flow.ReverseFlowPacketCount = *f.ReverseFlowPacketCount

	return flow, nil
}

type PostFlowsInput struct {
	Flows []FlowModel `json:"flows" binding:"required,dive"`
}

type PostFlowsOutput struct {
	Error          string `json:"error,omitempty"`
	ProcessedFlows int    `json:"processed_flows,omitempty"`
}

func PostFlows(enricher *enrichment.EnrichmentManager, outputStream *output.KinesisStream) gin.HandlerFunc {
	return func(c *gin.Context) {
		var flowsRequest PostFlowsInput
		err := c.ShouldBind(&flowsRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, PostFlowsOutput{Error: err.Error()})
			return
		}

		numParsedFlows := 0

		for _, flow := range flowsRequest.Flows {
			parsedFlow, err := flow.ToFlow()
			if err != nil {
				c.JSON(http.StatusBadRequest, PostFlowsOutput{Error: err.Error()})
				return
			}

			enricher.Enrich(parsedFlow)

			err = outputStream.AddFlow(parsedFlow)
			if err != nil {
				c.JSON(http.StatusBadRequest, PostFlowsOutput{Error: err.Error()})
				return
			}

			numParsedFlows += 1
		}
		c.JSON(http.StatusOK, PostFlowsOutput{ProcessedFlows: numParsedFlows})
	}
}

func AddFlowsRoutes(rg *gin.RouterGroup, enricher *enrichment.EnrichmentManager, outputStream *output.KinesisStream) {
	flows := rg.Group("/flows")

	flows.POST("/", PostFlows(enricher, outputStream))
}
