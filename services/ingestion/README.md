# flowsys ingestion service

The ingestion service is responsible for receiving data from various sources (API calls, IPFIX exporters, cloud log streams) and performing data enrichment and normalization before sending the data to a downstream Kinesis data stream.
