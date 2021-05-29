package enrichment

import (
	"fmt"
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/sheacloud/flowsys/services/ingestion/internal/schema"
)

type GeoIPEnricher struct {
	Language string
	db       *geoip2.Reader
}

func (e *GeoIPEnricher) Initialize() {
	db, err := geoip2.Open("./geoip-database/GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err)
		return
	}
	e.db = db
}

func (e *GeoIPEnricher) FlattenCity(city *geoip2.City) *schema.GeoIPData {
	metadata := &schema.GeoIPData{}
	if city.City.Names != nil {
		metadata.CityName = city.City.Names[e.Language]
	}
	if city.Continent.Names != nil {
		metadata.ContinentCode = city.Continent.Code
		metadata.ContinentName = city.Continent.Names[e.Language]
	}
	if city.Country.Names != nil {
		metadata.CountryIsoCode = city.Country.IsoCode
		metadata.CountryName = city.Country.Names[e.Language]
		metadata.CountryInEU = city.Country.IsInEuropeanUnion
	}
	if city.Location.TimeZone != "" {
		metadata.Latitude = city.Location.Latitude
		metadata.Longitude = city.Location.Longitude
		metadata.MetroCode = uint32(city.Location.MetroCode)
		metadata.TimeZone = city.Location.TimeZone
	}
	if city.Postal.Code != "" {
		metadata.PostalCode = city.Postal.Code
	}
	subdivisions := []*schema.Subdivision{}
	for _, sub := range city.Subdivisions {
		subdivision := &schema.Subdivision{
			IsoCode: sub.IsoCode,
			Name:    sub.Names[e.Language],
		}
		subdivisions = append(subdivisions, subdivision)
	}
	if subdivisions != nil {
		metadata.Subdivisions = subdivisions
	}

	return metadata
}

func (e *GeoIPEnricher) Enrich(flow *schema.Flow) {
	srcCityData, _ := e.db.City(net.IP(flow.GetSourceIPv4Address()))
	dstCityData, _ := e.db.City(net.IP(flow.GetDestinationIPv4Address()))

	flow.SourceGeoIPData = e.FlattenCity(srcCityData)
	flow.DestinationGeoIPData = e.FlattenCity(dstCityData)
}

func (e *GeoIPEnricher) GetName() string {
	return "GeoIPEnricher"
}
