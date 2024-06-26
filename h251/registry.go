// Code generated by "hl7fetch -pkgdir h251 -root ./genjson -version 2.5.1"; DO NOT EDIT.

// Package h251 contains the data structures for HL7 v2.5.1.
package h251

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
var Version = `2.5.1`

// Segments specific to file and batch control.
var ControlSegmentRegistry = map[string]any{
	"BHS": BHS{},
	"BTS": BTS{},
	"FHS": FHS{},
	"FTS": FTS{},
	"DSC": DSC{},
	"OVR": OVR{},
	"ADD": ADD{},
	"SFT": SFT{},
}

// Segment lookup by ID.
var SegmentRegistry = map[string]any{
	"ABS": ABS{},
	"ACC": ACC{},
	"ADD": ADD{},
	"AFF": AFF{},
	"AIG": AIG{},
	"AIL": AIL{},
	"AIP": AIP{},
	"AIS": AIS{},
	"AL1": AL1{},
	"APR": APR{},
	"ARQ": ARQ{},
	"AUT": AUT{},
	"BHS": BHS{},
	"BLC": BLC{},
	"BLG": BLG{},
	"BPO": BPO{},
	"BPX": BPX{},
	"BTS": BTS{},
	"BTX": BTX{},
	"CDM": CDM{},
	"CER": CER{},
	"CM0": CM0{},
	"CM1": CM1{},
	"CM2": CM2{},
	"CNS": CNS{},
	"CSP": CSP{},
	"CSR": CSR{},
	"CSS": CSS{},
	"CTD": CTD{},
	"CTI": CTI{},
	"DB1": DB1{},
	"DG1": DG1{},
	"DRG": DRG{},
	"DSC": DSC{},
	"DSP": DSP{},
	"ECD": ECD{},
	"ECR": ECR{},
	"EDU": EDU{},
	"EQL": EQL{},
	"EQP": EQP{},
	"EQU": EQU{},
	"ERQ": ERQ{},
	"ERR": ERR{},
	"EVN": EVN{},
	"FAC": FAC{},
	"FHS": FHS{},
	"FT1": FT1{},
	"FTS": FTS{},
	"GOL": GOL{},
	"GP1": GP1{},
	"GP2": GP2{},
	"GT1": GT1{},
	"IAM": IAM{},
	"IIM": IIM{},
	"IN1": IN1{},
	"IN2": IN2{},
	"IN3": IN3{},
	"INV": INV{},
	"IPC": IPC{},
	"ISD": ISD{},
	"LAN": LAN{},
	"LCC": LCC{},
	"LCH": LCH{},
	"LDP": LDP{},
	"LOC": LOC{},
	"LRL": LRL{},
	"MFA": MFA{},
	"MFE": MFE{},
	"MFI": MFI{},
	"MRG": MRG{},
	"MSA": MSA{},
	"MSH": MSH{},
	"NCK": NCK{},
	"NDS": NDS{},
	"NK1": NK1{},
	"NPU": NPU{},
	"NSC": NSC{},
	"NST": NST{},
	"NTE": NTE{},
	"OBR": OBR{},
	"OBX": OBX{},
	"ODS": ODS{},
	"ODT": ODT{},
	"OM1": OM1{},
	"OM2": OM2{},
	"OM3": OM3{},
	"OM4": OM4{},
	"OM5": OM5{},
	"OM6": OM6{},
	"OM7": OM7{},
	"ORC": ORC{},
	"ORG": ORG{},
	"OVR": OVR{},
	"PCR": PCR{},
	"PD1": PD1{},
	"PDA": PDA{},
	"PDC": PDC{},
	"PEO": PEO{},
	"PES": PES{},
	"PID": PID{},
	"PR1": PR1{},
	"PRA": PRA{},
	"PRB": PRB{},
	"PRC": PRC{},
	"PRD": PRD{},
	"PSH": PSH{},
	"PTH": PTH{},
	"PV1": PV1{},
	"PV2": PV2{},
	"QAK": QAK{},
	"QID": QID{},
	"QPD": QPD{},
	"QRD": QRD{},
	"QRF": QRF{},
	"QRI": QRI{},
	"RCP": RCP{},
	"RDF": RDF{},
	"RDT": RDT{},
	"RF1": RF1{},
	"RGS": RGS{},
	"RMI": RMI{},
	"ROL": ROL{},
	"RQ1": RQ1{},
	"RQD": RQD{},
	"RXA": RXA{},
	"RXC": RXC{},
	"RXD": RXD{},
	"RXE": RXE{},
	"RXG": RXG{},
	"RXO": RXO{},
	"RXR": RXR{},
	"SAC": SAC{},
	"SCH": SCH{},
	"SFT": SFT{},
	"SID": SID{},
	"SPM": SPM{},
	"SPR": SPR{},
	"STF": STF{},
	"TCC": TCC{},
	"TCD": TCD{},
	"TQ1": TQ1{},
	"TQ2": TQ2{},
	"TXA": TXA{},
	"UB1": UB1{},
	"UB2": UB2{},
	"URD": URD{},
	"URS": URS{},
	"VAR": VAR{},
	"VTQ": VTQ{},
}

// Trigger lookup by ID.
var TriggerRegistry = map[string]any{
	"ACK":     ACK{},
	"ADR_A19": ADR_A19{},
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
	"ADT_A25": ADT_A25{},
	"ADT_A26": ADT_A26{},
	"ADT_A27": ADT_A27{},
	"ADT_A28": ADT_A28{},
	"ADT_A29": ADT_A29{},
	"ADT_A30": ADT_A30{},
	"ADT_A31": ADT_A31{},
	"ADT_A32": ADT_A32{},
	"ADT_A33": ADT_A33{},
	"ADT_A34": ADT_A34{},
	"ADT_A35": ADT_A35{},
	"ADT_A36": ADT_A36{},
	"ADT_A37": ADT_A37{},
	"ADT_A38": ADT_A38{},
	"ADT_A39": ADT_A39{},
	"ADT_A40": ADT_A40{},
	"ADT_A41": ADT_A41{},
	"ADT_A42": ADT_A42{},
	"ADT_A43": ADT_A43{},
	"ADT_A44": ADT_A44{},
	"ADT_A45": ADT_A45{},
	"ADT_A46": ADT_A46{},
	"ADT_A47": ADT_A47{},
	"ADT_A48": ADT_A48{},
	"ADT_A49": ADT_A49{},
	"ADT_A50": ADT_A50{},
	"ADT_A51": ADT_A51{},
	"ADT_A52": ADT_A52{},
	"ADT_A53": ADT_A53{},
	"ADT_A54": ADT_A54{},
	"ADT_A55": ADT_A55{},
	"ADT_A60": ADT_A60{},
	"ADT_A61": ADT_A61{},
	"ADT_A62": ADT_A62{},
	"BAR_P01": BAR_P01{},
	"BAR_P02": BAR_P02{},
	"BAR_P05": BAR_P05{},
	"BAR_P06": BAR_P06{},
	"BAR_P10": BAR_P10{},
	"BAR_P12": BAR_P12{},
	"BPS_O29": BPS_O29{},
	"BRP_O30": BRP_O30{},
	"BRT_O32": BRT_O32{},
	"BTS_O31": BTS_O31{},
	"CRM_C01": CRM_C01{},
	"CRM_C02": CRM_C02{},
	"CRM_C03": CRM_C03{},
	"CRM_C04": CRM_C04{},
	"CRM_C05": CRM_C05{},
	"CRM_C06": CRM_C06{},
	"CRM_C07": CRM_C07{},
	"CRM_C08": CRM_C08{},
	"CSU_C09": CSU_C09{},
	"CSU_C10": CSU_C10{},
	"CSU_C11": CSU_C11{},
	"CSU_C12": CSU_C12{},
	"DFT_P03": DFT_P03{},
	"DFT_P11": DFT_P11{},
	"DOC_T12": DOC_T12{},
	"DSR_Q01": DSR_Q01{},
	"DSR_Q03": DSR_Q03{},
	"EAC_U07": EAC_U07{},
	"EAN_U09": EAN_U09{},
	"EAR_U08": EAR_U08{},
	"EDR_R07": EDR_R07{},
	"EQQ_Q04": EQQ_Q04{},
	"ERP_R09": ERP_R09{},
	"ESR_U02": ESR_U02{},
	"ESU_U01": ESU_U01{},
	"INR_U06": INR_U06{},
	"INU_U05": INU_U05{},
	"LSR_U13": LSR_U13{},
	"LSU_U12": LSU_U12{},
	"MDM_T01": MDM_T01{},
	"MDM_T02": MDM_T02{},
	"MDM_T03": MDM_T03{},
	"MDM_T04": MDM_T04{},
	"MDM_T05": MDM_T05{},
	"MDM_T06": MDM_T06{},
	"MDM_T07": MDM_T07{},
	"MDM_T08": MDM_T08{},
	"MDM_T09": MDM_T09{},
	"MDM_T10": MDM_T10{},
	"MDM_T11": MDM_T11{},
	"MFK_M01": MFK_M01{},
	"MFK_M02": MFK_M02{},
	"MFK_M04": MFK_M04{},
	"MFN_M02": MFN_M02{},
	"MFN_M04": MFN_M04{},
	"MFN_M05": MFN_M05{},
	"MFN_M06": MFN_M06{},
	"MFN_M07": MFN_M07{},
	"MFN_M08": MFN_M08{},
	"MFN_M09": MFN_M09{},
	"MFN_M10": MFN_M10{},
	"MFN_M11": MFN_M11{},
	"MFN_M12": MFN_M12{},
	"MFN_M13": MFN_M13{},
	"MFN_M15": MFN_M15{},
	"MFQ_M01": MFQ_M01{},
	"MFR_M04": MFR_M04{},
	"MFR_M05": MFR_M05{},
	"MFR_M06": MFR_M06{},
	"MFR_M07": MFR_M07{},
	"NMD_N02": NMD_N02{},
	"NMQ_N01": NMQ_N01{},
	"NMR_N01": NMR_N01{},
	"OMB_O27": OMB_O27{},
	"OMD_O03": OMD_O03{},
	"OMG_O19": OMG_O19{},
	"OMI_O23": OMI_O23{},
	"OML_O21": OML_O21{},
	"OML_O33": OML_O33{},
	"OML_O35": OML_O35{},
	"OMN_O07": OMN_O07{},
	"OMP_O09": OMP_O09{},
	"OMS_O05": OMS_O05{},
	"ORB_O28": ORB_O28{},
	"ORD_O04": ORD_O04{},
	"ORF_R04": ORF_R04{},
	"ORF_W02": ORF_W02{},
	"ORG_O20": ORG_O20{},
	"ORI_O24": ORI_O24{},
	"ORL_O22": ORL_O22{},
	"ORL_O34": ORL_O34{},
	"ORL_O36": ORL_O36{},
	"ORM_O01": ORM_O01{},
	"ORN_O08": ORN_O08{},
	"ORP_O10": ORP_O10{},
	"ORR_O02": ORR_O02{},
	"ORS_O06": ORS_O06{},
	"ORU_R01": ORU_R01{},
	"ORU_R30": ORU_R30{},
	"ORU_R31": ORU_R31{},
	"ORU_R32": ORU_R32{},
	"ORU_W01": ORU_W01{},
	"OSQ_Q06": OSQ_Q06{},
	"OSR_Q06": OSR_Q06{},
	"OUL_R21": OUL_R21{},
	"OUL_R22": OUL_R22{},
	"OUL_R23": OUL_R23{},
	"OUL_R24": OUL_R24{},
	"PEX_P07": PEX_P07{},
	"PEX_P08": PEX_P08{},
	"PGL_PC6": PGL_PC6{},
	"PGL_PC7": PGL_PC7{},
	"PGL_PC8": PGL_PC8{},
	"PIN_I07": PIN_I07{},
	"PMU_B01": PMU_B01{},
	"PMU_B02": PMU_B02{},
	"PMU_B03": PMU_B03{},
	"PMU_B04": PMU_B04{},
	"PMU_B05": PMU_B05{},
	"PMU_B06": PMU_B06{},
	"PMU_B07": PMU_B07{},
	"PMU_B08": PMU_B08{},
	"PPG_PCG": PPG_PCG{},
	"PPG_PCH": PPG_PCH{},
	"PPG_PCJ": PPG_PCJ{},
	"PPP_PCB": PPP_PCB{},
	"PPP_PCC": PPP_PCC{},
	"PPP_PCD": PPP_PCD{},
	"PPR_PC1": PPR_PC1{},
	"PPR_PC2": PPR_PC2{},
	"PPR_PC3": PPR_PC3{},
	"PPT_PCL": PPT_PCL{},
	"PPV_PCA": PPV_PCA{},
	"PRR_PC5": PRR_PC5{},
	"PTR_PCF": PTR_PCF{},
	"QBP_Q11": QBP_Q11{},
	"QBP_Q13": QBP_Q13{},
	"QBP_Q15": QBP_Q15{},
	"QBP_Q21": QBP_Q21{},
	"QBP_Q22": QBP_Q22{},
	"QBP_Q23": QBP_Q23{},
	"QBP_Q24": QBP_Q24{},
	"QBP_Q25": QBP_Q25{},
	"QBP_Q31": QBP_Q31{},
	"QBP_Z73": QBP_Z73{},
	"QBP_Z75": QBP_Z75{},
	"QBP_Z77": QBP_Z77{},
	"QBP_Z79": QBP_Z79{},
	"QBP_Z81": QBP_Z81{},
	"QBP_Z85": QBP_Z85{},
	"QBP_Z87": QBP_Z87{},
	"QBP_Z89": QBP_Z89{},
	"QBP_Z91": QBP_Z91{},
	"QBP_Z93": QBP_Z93{},
	"QBP_Z95": QBP_Z95{},
	"QBP_Z97": QBP_Z97{},
	"QBP_Z99": QBP_Z99{},
	"QCK_Q02": QCK_Q02{},
	"QCN_J01": QCN_J01{},
	"QRY_A19": QRY_A19{},
	"QRY_PC4": QRY_PC4{},
	"QRY_PC9": QRY_PC9{},
	"QRY_PCE": QRY_PCE{},
	"QRY_PCK": QRY_PCK{},
	"QRY_Q01": QRY_Q01{},
	"QRY_Q02": QRY_Q02{},
	"QRY_Q26": QRY_Q26{},
	"QRY_Q27": QRY_Q27{},
	"QRY_Q28": QRY_Q28{},
	"QRY_Q29": QRY_Q29{},
	"QRY_Q30": QRY_Q30{},
	"QRY_R02": QRY_R02{},
	"QRY_T12": QRY_T12{},
	"QSB_Q16": QSB_Q16{},
	"QSB_Z83": QSB_Z83{},
	"QSX_J02": QSX_J02{},
	"QVR_Q17": QVR_Q17{},
	"RAR_RAR": RAR_RAR{},
	"RAS_O17": RAS_O17{},
	"RCI_I05": RCI_I05{},
	"RCL_I06": RCL_I06{},
	"RDE_O11": RDE_O11{},
	"RDE_O25": RDE_O25{},
	"RDR_RDR": RDR_RDR{},
	"RDS_O13": RDS_O13{},
	"RDY_K15": RDY_K15{},
	"RDY_Z80": RDY_Z80{},
	"RDY_Z98": RDY_Z98{},
	"REF_I12": REF_I12{},
	"REF_I13": REF_I13{},
	"REF_I14": REF_I14{},
	"REF_I15": REF_I15{},
	"RER_RER": RER_RER{},
	"RGR_RGR": RGR_RGR{},
	"RGV_O15": RGV_O15{},
	"ROR_ROR": ROR_ROR{},
	"RPA_I08": RPA_I08{},
	"RPA_I09": RPA_I09{},
	"RPA_I10": RPA_I10{},
	"RPA_I11": RPA_I11{},
	"RPI_I01": RPI_I01{},
	"RPI_I04": RPI_I04{},
	"RPL_I02": RPL_I02{},
	"RPR_I03": RPR_I03{},
	"RQA_I08": RQA_I08{},
	"RQA_I09": RQA_I09{},
	"RQA_I10": RQA_I10{},
	"RQA_I11": RQA_I11{},
	"RQC_I05": RQC_I05{},
	"RQC_I06": RQC_I06{},
	"RQI_I01": RQI_I01{},
	"RQI_I02": RQI_I02{},
	"RQI_I03": RQI_I03{},
	"RQP_I04": RQP_I04{},
	"RQQ_Q09": RQQ_Q09{},
	"RRA_O18": RRA_O18{},
	"RRD_O14": RRD_O14{},
	"RRE_O12": RRE_O12{},
	"RRE_O26": RRE_O26{},
	"RRG_O16": RRG_O16{},
	"RRI_I12": RRI_I12{},
	"RRI_I13": RRI_I13{},
	"RRI_I14": RRI_I14{},
	"RRI_I15": RRI_I15{},
	"RSP_K11": RSP_K11{},
	"RSP_K21": RSP_K21{},
	"RSP_K22": RSP_K22{},
	"RSP_K23": RSP_K23{},
	"RSP_K24": RSP_K24{},
	"RSP_K25": RSP_K25{},
	"RSP_K31": RSP_K31{},
	"RSP_Z82": RSP_Z82{},
	"RSP_Z84": RSP_Z84{},
	"RSP_Z86": RSP_Z86{},
	"RSP_Z88": RSP_Z88{},
	"RSP_Z90": RSP_Z90{},
	"RTB_K13": RTB_K13{},
	"RTB_Z74": RTB_Z74{},
	"RTB_Z76": RTB_Z76{},
	"RTB_Z78": RTB_Z78{},
	"RTB_Z92": RTB_Z92{},
	"RTB_Z94": RTB_Z94{},
	"RTB_Z96": RTB_Z96{},
	"SIU_S12": SIU_S12{},
	"SIU_S13": SIU_S13{},
	"SIU_S14": SIU_S14{},
	"SIU_S15": SIU_S15{},
	"SIU_S16": SIU_S16{},
	"SIU_S17": SIU_S17{},
	"SIU_S18": SIU_S18{},
	"SIU_S19": SIU_S19{},
	"SIU_S20": SIU_S20{},
	"SIU_S21": SIU_S21{},
	"SIU_S22": SIU_S22{},
	"SIU_S23": SIU_S23{},
	"SIU_S24": SIU_S24{},
	"SIU_S26": SIU_S26{},
	"SPQ_Q08": SPQ_Q08{},
	"SQM_S25": SQM_S25{},
	"SQR_S25": SQR_S25{},
	"SRM_S01": SRM_S01{},
	"SRM_S02": SRM_S02{},
	"SRM_S03": SRM_S03{},
	"SRM_S04": SRM_S04{},
	"SRM_S05": SRM_S05{},
	"SRM_S06": SRM_S06{},
	"SRM_S07": SRM_S07{},
	"SRM_S08": SRM_S08{},
	"SRM_S09": SRM_S09{},
	"SRM_S10": SRM_S10{},
	"SRM_S11": SRM_S11{},
	"SRR_S01": SRR_S01{},
	"SRR_S02": SRR_S02{},
	"SRR_S03": SRR_S03{},
	"SRR_S04": SRR_S04{},
	"SRR_S05": SRR_S05{},
	"SRR_S06": SRR_S06{},
	"SRR_S07": SRR_S07{},
	"SRR_S08": SRR_S08{},
	"SRR_S09": SRR_S09{},
	"SRR_S10": SRR_S10{},
	"SRR_S11": SRR_S11{},
	"SSR_U04": SSR_U04{},
	"SSU_U03": SSU_U03{},
	"SUR_P09": SUR_P09{},
	"TBR_R08": TBR_R08{},
	"TCR_U11": TCR_U11{},
	"TCU_U10": TCU_U10{},
	"UDM_Q05": UDM_Q05{},
	"VQQ_Q07": VQQ_Q07{},
	"VXQ_V01": VXQ_V01{},
	"VXR_V03": VXR_V03{},
	"VXU_V04": VXU_V04{},
	"VXX_V02": VXX_V02{},
}

// Data Type lookup by ID.
var DataTypeRegistry = map[string]any{
	"AD":     *(new(AD)),
	"AUI":    *(new(AUI)),
	"CCD":    *(new(CCD)),
	"CE":     *(new(CE)),
	"CNE":    *(new(CNE)),
	"CNN":    *(new(CNN)),
	"CP":     *(new(CP)),
	"CQ":     *(new(CQ)),
	"CWE":    *(new(CWE)),
	"CX":     *(new(CX)),
	"DDI":    *(new(DDI)),
	"DIN":    *(new(DIN)),
	"DLD":    *(new(DLD)),
	"DLN":    *(new(DLN)),
	"DLT":    *(new(DLT)),
	"DR":     *(new(DR)),
	"DT":     *(new(DT)),
	"DTM":    *(new(DTM)),
	"DTN":    *(new(DTN)),
	"ED":     *(new(ED)),
	"EI":     *(new(EI)),
	"EIP":    *(new(EIP)),
	"ELD":    *(new(ELD)),
	"ERL":    *(new(ERL)),
	"FC":     *(new(FC)),
	"FN":     *(new(FN)),
	"FT":     *(new(FT)),
	"GTS":    *(new(GTS)),
	"HD":     *(new(HD)),
	"ICD":    *(new(ICD)),
	"ID":     *(new(ID)),
	"IS":     *(new(IS)),
	"JCC":    *(new(JCC)),
	"LA1":    *(new(LA1)),
	"LA2":    *(new(LA2)),
	"MO":     *(new(MO)),
	"MOC":    *(new(MOC)),
	"MOP":    *(new(MOP)),
	"MSG":    *(new(MSG)),
	"NA":     *(new(NA)),
	"NDL":    *(new(NDL)),
	"NM":     *(new(NM)),
	"NR":     *(new(NR)),
	"OCD":    *(new(OCD)),
	"OSD":    *(new(OSD)),
	"OSP":    *(new(OSP)),
	"PIP":    *(new(PIP)),
	"PL":     *(new(PL)),
	"PLN":    *(new(PLN)),
	"PPN":    *(new(PPN)),
	"PRL":    *(new(PRL)),
	"PT":     *(new(PT)),
	"PTA":    *(new(PTA)),
	"QIP":    *(new(QIP)),
	"QSC":    *(new(QSC)),
	"RCD":    *(new(RCD)),
	"RFR":    *(new(RFR)),
	"RI":     *(new(RI)),
	"RMC":    *(new(RMC)),
	"RPT":    *(new(RPT)),
	"SAD":    *(new(SAD)),
	"SCV":    *(new(SCV)),
	"SI":     *(new(SI)),
	"SN":     *(new(SN)),
	"SPD":    *(new(SPD)),
	"SPS":    *(new(SPS)),
	"SRT":    *(new(SRT)),
	"ST":     *(new(ST)),
	"TM":     *(new(TM)),
	"TQ":     *(new(TQ)),
	"TS":     *(new(TS)),
	"TX":     *(new(TX)),
	"UVC":    *(new(UVC)),
	"VARIES": *(new(VARIES)),
	"VH":     *(new(VH)),
	"VID":    *(new(VID)),
	"VR":     *(new(VR)),
	"XAD":    *(new(XAD)),
	"XCN":    *(new(XCN)),
	"XON":    *(new(XON)),
	"XPN":    *(new(XPN)),
	"XTN":    *(new(XTN)),
}
