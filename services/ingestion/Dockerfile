FROM golang:1.16-alpine as build

COPY . /build
RUN cd /build; CGO_ENABLED=1 GOBIN=/bin/ go install ./cmd/ingestion/;

FROM alpine as prod

COPY --from=build /bin/ingestion /bin/ingestion
COPY ./geoip-database/GeoLite2-City.mmdb /ingestion/geoip-database/

WORKDIR /ingestion/

CMD ["/bin/ingestion"]
