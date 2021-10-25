package schema

import (
	"encoding/json"
)

type FlowModel struct {
	SourceIPv4Address      *string `json:"source_ipv4_address"`
	DestinationIPv4Address *string `json:"destination_ipv4_address"`
	SourceIPv6Address      *string `json:"source_ipv6_address"`
	DestinationIPv6Address *string `json:"destination_ipv6_address"`
	// SourceIPv4AddressNumber      *uint32           `json:"source_ipv4_address_number"`
	// DestinationIPv4AddressNumber *uint32           `json:"destination_ipv4_address_number"`
	// SourceIPv6AddressNumber      *uint64           `json:"source_ipv6_address_number"`
	// DestinationIPv6AddressNumber *uint64           `json:"destination_ipv6_address_number"`
	SourcePort             uint16            `json:"source_port" binding:"required"`
	DestinationPort        uint16            `json:"destination_port" binding:"required"`
	IpProtocolVersion      IpProtocolVersion `json:"ip_protocol_version" binding:"required"`
	TransportProtocol      TransportProtocol `json:"transport_protocol" binding:"required"`
	FlowStartMilliseconds  uint64            `json:"flow_start_milliseconds" binding:"required"`
	FlowEndMilliseconds    uint64            `json:"flow_end_milliseconds" binding:"required"`
	FlowOctetCount         uint64            `json:"flow_octet_count" binding:"required"`
	FlowPacketCount        uint64            `json:"flow_packet_count" binding:"required"`
	ReverseFlowOctetCount  uint64            `json:"reverse_flow_octet_count"`
	ReverseFlowPacketCount uint64            `json:"reverse_flow_packet_count"`
}

type IpProtocolVersion uint8

var (
	IpProtocolVersion4 = IpProtocolVersion(4)
	IpProtocolVersion6 = IpProtocolVersion(6)
)

type TransportProtocol uint8

var (
	TransportProtocolTCP = TransportProtocol(6)
	TransportProtocolUDP = TransportProtocol(17)
)

// func ipv4ToNumber(ipString *string) *uint32 {
// 	if ipString == nil {
// 		return nil
// 	}

// 	ip := net.ParseIP(*ipString)
// 	if ip == nil {
// 		return nil
// 	}

// 	ipNum := ip.To4()
// 	if ipNum == nil {
// 		return nil
// 	}

// 	ipNumUint32 := uint32(ipNum[0])<<24 | uint32(ipNum[1])<<16 | uint32(ipNum[2])<<8 | uint32(ipNum[3])
// 	return &ipNumUint32
// }

// func ipv6ToNumber(ipString *string) *uint64 {
// 	if ipString == nil {
// 		return nil
// 	}

// 	ip := net.ParseIP(*ipString)
// 	if ip == nil {
// 		return nil
// 	}

// 	ipNum := ip.To16()
// 	if ipNum == nil {
// 		return nil
// 	}

// 	ipNumUint64 := uint64(ipNum[0])<<56 | uint64(ipNum[1])<<48 | uint64(ipNum[2])<<40 | uint64(ipNum[3])<<32 | uint64(ipNum[4])<<24 | uint64(ipNum[5])<<16 | uint64(ipNum[6])<<8 | uint64(ipNum[7])
// 	return &ipNumUint64
// }

func (f *FlowModel) PopulateFields() {
	// f.SourceIPv4AddressNumber = ipv4ToNumber(f.SourceIPv4Address)
	// f.DestinationIPv4AddressNumber = ipv4ToNumber(f.DestinationIPv4Address)
	// f.SourceIPv6AddressNumber = ipv6ToNumber(f.SourceIPv6Address)
	// f.DestinationIPv6AddressNumber = ipv6ToNumber(f.DestinationIPv6Address)
}

func (f *FlowModel) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}
