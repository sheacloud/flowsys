package enrichment

import (
	"github.com/sheacloud/flowsys/schema"
	"github.com/sirupsen/logrus"
)

type Enricher interface {
	Enrich(flow *schema.Flow)
	GetName() string
}

type EnrichmentManager struct {
	Enrichers []Enricher
}

func NewEnrichmentManager(enrichers []Enricher) EnrichmentManager {
	return EnrichmentManager{
		Enrichers: enrichers,
	}
}

func (em *EnrichmentManager) Enrich(flow *schema.Flow) {
	for _, enricher := range em.Enrichers {
		logrus.WithFields(logrus.Fields{
			"flow":     flow.String(),
			"enricher": enricher.GetName(),
		}).Trace("Enriching flow")
		enricher.Enrich(flow)
	}
}
