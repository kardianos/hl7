package main

type resource string

const (
	ResourceChapters      resource = "Chapters"
	ResourceTriggerEvents resource = "TriggerEvents"
	ResourceSegments      resource = "Segments"
	ResourceDataTypes     resource = "DataTypes"
	ResourceTables        resource = "Tables"
)

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/DataTypes/CE
type DataType struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	DataType     interface{} `json:"dataType"`
	DataTypeName interface{} `json:"dataTypeName"`
	Length       int         `json:"length"`
	Usage        interface{} `json:"usage"`
	Rpt          interface{} `json:"rpt"`
	TableID      interface{} `json:"tableId"`
	TableName    interface{} `json:"tableName"`
	Sample       string      `json:"sample"`
	Fields       []Field     `json:"fields"`
}

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Segments/NK1
type SegmentType struct {
	ID          string   `json:"id"`
	SegmentID   string   `json:"segmentId"`
	LongName    string   `json:"longName"`
	Description string   `json:"description"`
	Sample      string   `json:"sample"`
	Chapters    []string `json:"chapters"`
	Fields      []Field  `json:"fields"`
}
type Field struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Position     string `json:"position"`
	Length       int    `json:"length"`
	DataType     string `json:"dataType"`
	DataTypeName string `json:"dataTypeName"`
	Usage        string `json:"usage"`
	Rpt          string `json:"rpt"`
	TableID      string `json:"tableId"`
	TableName    string `json:"tableName"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Tables/0052
type TableType struct {
	ID        string     `json:"id"`
	TableID   string     `json:"tableId"`
	TableType string     `json:"tableType"`
	Name      string     `json:"name"`
	Chapters  []string   `json:"chapters"`
	Entries   []TableRow `json:"entries"`
}
type TableRow struct {
	Value       string `json:"value"`
	Description string `json:"description"`
	Comment     string `json:"comment"`
}

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/TriggerEvents/ORU_R30
type TriggerType struct {
	ID          string           `json:"id"`
	MsgStructID string           `json:"msgStructId"`
	EventDesc   string           `json:"eventDesc"`
	Description string           `json:"description"`
	Sample      interface{}      `json:"sample"`
	Chapters    []string         `json:"chapters"`
	Segments    []TriggerSegment `json:"segments"`
}
type TriggerSegment struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	LongName string           `json:"longName"`
	Sequence string           `json:"sequence"`
	Usage    string           `json:"usage"`
	Rpt      string           `json:"rpt"`
	IsGroup  bool             `json:"isGroup"`
	Segments []TriggerSegment `json:"segments"`
}

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/TriggerEvents
// []TriggerEvents
type TriggerEvents struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Chapters    []string `json:"chapters"`
}

// https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Chapters
// []Chapter
type Chapter struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
