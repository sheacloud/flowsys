package enrichment

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/sheacloud/flowsys/schema"
	"github.com/sirupsen/logrus"
)

var (
	notFound = ttlcache.ErrNotFound
)

type DNSEnricher struct {
	timeout  int
	cache    ttlcache.SimpleCache
	resolver net.Resolver
}

func NewDNSEnricher(lookupTimeout, cacheTimeout int) *DNSEnricher {
	cache := ttlcache.NewCache()
	cache.SetTTL(time.Duration(cacheTimeout) * time.Minute)

	return &DNSEnricher{
		timeout:  lookupTimeout,
		cache:    cache,
		resolver: net.Resolver{},
	}
}

func (d *DNSEnricher) LookupReverse(ip net.IP) (string, error) {
	if val, err := d.cache.Get(ip.String()); err != notFound {
		logrus.WithFields(logrus.Fields{
			"ip":   ip.String(),
			"name": val.(string),
		}).Info("Got reverse DNS name from cache")
		return val.(string), nil
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(d.timeout)*time.Second)
	defer cancel()
	names, err := d.resolver.LookupAddr(ctx, ip.String())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error performing reverse lookup")
	}

	if len(names) > 0 {
		d.cache.Set(ip.String(), names[0])
		logrus.WithFields(logrus.Fields{
			"ip":   ip.String(),
			"name": names[0],
		}).Info("Got reverse DNS name")
		return names[0], nil
	} else {
		return "", fmt.Errorf("no reverse DNS name found")
	}
}

func (d *DNSEnricher) Enrich(flow *schema.Flow) {
	sourceDomain, err := d.LookupReverse(flow.GetSourceIPv4Address())
	if err == nil {
		flow.SourceDomainName = sourceDomain
	}

	destinationDomain, err := d.LookupReverse(flow.GetDestinationIPv4Address())
	if err == nil {
		flow.DestinationDomainName = destinationDomain
	}
}

func (d *DNSEnricher) GetName() string {
	return "DNSEnricher"
}
