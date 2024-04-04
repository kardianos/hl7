// Code generated by "hl7fetch -pkgdir h210 -root ./genjson -version 2.1"; DO NOT EDIT.

// Package h210 contains the data structures for HL7 v2.1.
package h210

// Registry implements the required interface for unmarshalling data.
var Registry = registry{}

type registry struct{}

func (registry) Version() string {
	return Version
}
func (registry) ControlSegment(name string) (any, bool) {
	v, ok := ControlSegmentRegistry[name]
	return v, ok
}
func (registry) Segment(name string) (any, bool) {
	v, ok := SegmentRegistry[name]
	return v, ok
}
func (registry) Trigger(name string) (any, bool) {
	v, ok := TriggerRegistry[name]
	return v, ok
}
func (registry) DataType(name string) (any, bool) {
	v, ok := DataTypeRegistry[name]
	return v, ok
}

// Version of this HL7 package.
var Version = `2.1`

// Segments specific to file and batch control.
var ControlSegmentRegistry = map[string]any{
	"BHS": BHS{},
	"BTS": BTS{},
	"FHS": FHS{},
	"FTS": FTS{},
	"DSC": DSC{},
	"ADD": ADD{},
}

// Segment lookup by ID.
var SegmentRegistry = map[string]any{
	"ACC": ACC{},
	"ADD": ADD{},
	"BHS": BHS{},
	"BLG": BLG{},
	"BTS": BTS{},
	"DG1": DG1{},
	"DSC": DSC{},
	"DSP": DSP{},
	"EVN": EVN{},
	"FHS": FHS{},
	"FT1": FT1{},
	"FTS": FTS{},
	"GT1": GT1{},
	"IN1": IN1{},
	"MRG": MRG{},
	"MSA": MSA{},
	"MSH": MSH{},
	"NK1": NK1{},
	"NPU": NPU{},
	"NTE": NTE{},
	"OBR": OBR{},
	"OBX": OBX{},
	"ORC": ORC{},
	"PD1": PD1{},
	"PID": PID{},
	"PR1": PR1{},
	"PV1": PV1{},
	"QRD": QRD{},
	"QRF": QRF{},
	"UB1": UB1{},
	"URD": URD{},
	"URS": URS{},
}

// Trigger lookup by ID.
var TriggerRegistry = map[string]any{
	"ADT_A01": ADT_A01{},
	"ADT_A02": ADT_A02{},
	"ADT_A03": ADT_A03{},
	"ADT_A04": ADT_A04{},
	"ADT_A05": ADT_A05{},
	"ADT_A06": ADT_A06{},
	"ADT_A07": ADT_A07{},
	"ADT_A08": ADT_A08{},
	"ADT_A09": ADT_A09{},
	"ADT_A10": ADT_A10{},
	"ADT_A11": ADT_A11{},
	"ADT_A12": ADT_A12{},
	"ADT_A13": ADT_A13{},
	"ADT_A14": ADT_A14{},
	"ADT_A15": ADT_A15{},
	"ADT_A16": ADT_A16{},
	"ADT_A17": ADT_A17{},
	"ADT_A18": ADT_A18{},
	"ADT_A20": ADT_A20{},
	"ADT_A21": ADT_A21{},
	"ADT_A22": ADT_A22{},
	"ADT_A23": ADT_A23{},
	"ADT_A24": ADT_A24{},
	"BAR_P01": BAR_P01{},
	"BAR_P02": BAR_P02{},
	"DFT_P03": DFT_P03{},
	"DSR_Q03": DSR_Q03{},
	"ORM_O01": ORM_O01{},
	"ORR_O02": ORR_O02{},
	"ORU_R01": ORU_R01{},
	"ORU_R03": ORU_R03{},
	"QRY_A19": QRY_A19{},
	"QRY_Q01": QRY_Q01{},
	"QRY_Q02": QRY_Q02{},
	"UDM_Q05": UDM_Q05{},
}

// Data Type lookup by ID.
var DataTypeRegistry = map[string]any{
	"AD": *(new(AD)),
	"CE": *(new(CE)),
	"CK": *(new(CK)),
	"CM": *(new(CM)),
	"CN": *(new(CN)),
	"CQ": *(new(CQ)),
	"DT": *(new(DT)),
	"ID": *(new(ID)),
	"NM": *(new(NM)),
	"PN": *(new(PN)),
	"SI": *(new(SI)),
	"ST": *(new(ST)),
	"TN": *(new(TN)),
	"TS": *(new(TS)),
	"TX": *(new(TX)),
}
