package hl7

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	v25 "github.com/kardianos/hl7/h250"
)

// TestComplexTriggerRoundTrip tests encoding and decoding of complex triggers
// with nested groups, repeated segments, and optional fields.
func TestComplexTriggerRoundTrip(t *testing.T) {
	t.Run("ORU_R01_SinglePatientMultipleObservations", testORU_R01_SinglePatientMultipleObservations)
	t.Run("ORU_R01_MultiplePatients", testORU_R01_MultiplePatients)
	t.Run("ORU_R01_DeepNesting", testORU_R01_DeepNesting)
	t.Run("ORL_O34_NestedSpecimens", testORL_O34_NestedSpecimens)
	t.Run("ORU_R01_MultipleOrderObservations", testORU_R01_MultipleOrderObservations)
}

func testORU_R01_SinglePatientMultipleObservations(t *testing.T) {
	// Create an ORU_R01 with one patient and multiple observations
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			SendingFacility:    &v25.HD{NamespaceID: "HOSPITAL"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG001",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Doe", GivenName: "John"},
						},
						DateTimeOfBirth:      time.Date(1980, 5, 15, 0, 0, 0, 0, time.UTC),
						AdministrativeSex:    "M",
						PatientAccountNumber: &v25.CX{IDNumber: "ACC123"},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                        "1",
							PlacerOrderNumber:            &v25.EI{EntityIdentifier: "ORDER001"},
							UniversalServiceIdentifier:   v25.CE{Identifier: "GLU", Text: "Glucose"},
							ObservationDateTime:          time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC),
							ResultsRptStatusChngDateTime: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                    "1",
									ValueType:                "NM",
									ObservationIdentifier:    v25.CE{Identifier: "GLU", Text: "Glucose"},
									Units:                    &v25.CE{Identifier: "mg/dL"},
									ObservationResultStatus:  "F",
									DateTimeOfTheObservation: time.Date(2025, 1, 15, 9, 30, 0, 0, time.UTC),
								},
							},
							{
								OBX: &v25.OBX{
									SetID:                    "2",
									ValueType:                "NM",
									ObservationIdentifier:    v25.CE{Identifier: "HBA1C", Text: "Hemoglobin A1c"},
									Units:                    &v25.CE{Identifier: "%"},
									ObservationResultStatus:  "F",
									DateTimeOfTheObservation: time.Date(2025, 1, 15, 9, 35, 0, 0, time.UTC),
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testORU_R01_MultiplePatients(t *testing.T) {
	// Create an ORU_R01 with multiple patients
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG002",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Smith", GivenName: "Jane"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "CBC"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "WBC"},
								},
							},
						},
					},
				},
			},
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Johnson", GivenName: "Bob"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "BMP"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "NA"},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testORU_R01_DeepNesting(t *testing.T) {
	// Test deep nesting with Visit, multiple NK1, NTE segments
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG003",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Wilson", GivenName: "Mary"},
						},
					},
					NTE: []v25.NTE{
						{SetID: "1", Comment: []string{"Patient note 1"}},
						{SetID: "2", Comment: []string{"Patient note 2"}},
					},
					NK1: []v25.NK1{
						{SetID: "1", NKName: []v25.XPN{{FamilyName: "Wilson", GivenName: "Tom"}}},
						{SetID: "2", NKName: []v25.XPN{{FamilyName: "Wilson", GivenName: "Sue"}}},
					},
					Visit: &v25.ORU_R01_Visit{
						PV1: &v25.PV1{
							SetID:        "1",
							PatientClass: "I",
							AssignedPatientLocation: &v25.PL{
								PointOfCare: "ICU",
								Room:        "101",
								Bed:         "A",
							},
						},
						PV2: &v25.PV2{
							AdmitReason: &v25.CE{Text: "Chest pain"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "TROPONIN"},
						},
						NTE: []v25.NTE{
							{SetID: "1", Comment: []string{"Order note"}},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "TROP-I"},
								},
								NTE: []v25.NTE{
									{SetID: "1", Comment: []string{"OBX note 1"}},
									{SetID: "2", Comment: []string{"OBX note 2"}},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testORL_O34_NestedSpecimens(t *testing.T) {
	// Test ORL_O34 with nested specimens and orders
	msg := v25.ORL_O34{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORL",
				TriggerEvent:     "O34",
				MessageStructure: "ORL_O34",
			},
			MessageControlID: "MSG004",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		MSA: &v25.MSA{
			AcknowledgmentCode: "AA",
			MessageControlID:   "REF001",
		},
		Response: &v25.ORL_O34_Response{
			Patient: &v25.ORL_O34_Patient{
				PID: &v25.PID{
					SetID: "1",
					PatientName: []v25.XPN{
						{FamilyName: "TestPatient", GivenName: "Lab"},
					},
				},
				Specimen: []v25.ORL_O34_Specimen{
					{
						SPM: &v25.SPM{
							SetID:        "1",
							SpecimenType: v25.CWE{Identifier: "BLOOD", Text: "Blood"},
						},
						Order: []v25.ORL_O34_Order{
							{
								ORC: &v25.ORC{
									OrderControl:      "OK",
									PlacerOrderNumber: &v25.EI{EntityIdentifier: "PO001"},
								},
								ObservationRequest: &v25.ORL_O34_ObservationRequest{
									OBR: &v25.OBR{
										SetID:                      "1",
										UniversalServiceIdentifier: v25.CE{Identifier: "GLU"},
									},
								},
							},
							{
								ORC: &v25.ORC{
									OrderControl:      "OK",
									PlacerOrderNumber: &v25.EI{EntityIdentifier: "PO002"},
								},
								ObservationRequest: &v25.ORL_O34_ObservationRequest{
									OBR: &v25.OBR{
										SetID:                      "2",
										UniversalServiceIdentifier: v25.CE{Identifier: "BUN"},
									},
								},
							},
						},
					},
					{
						SPM: &v25.SPM{
							SetID:        "2",
							SpecimenType: v25.CWE{Identifier: "URINE", Text: "Urine"},
						},
						Order: []v25.ORL_O34_Order{
							{
								ORC: &v25.ORC{
									OrderControl:      "OK",
									PlacerOrderNumber: &v25.EI{EntityIdentifier: "PO003"},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testORU_R01_MultipleOrderObservations(t *testing.T) {
	// Test multiple order observations with ORC
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG005",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Brown", GivenName: "Alice"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						ORC: &v25.ORC{
							OrderControl:      "RE",
							PlacerOrderNumber: &v25.EI{EntityIdentifier: "ORD001"},
							FillerOrderNumber: &v25.EI{EntityIdentifier: "FILL001"},
						},
						OBR: &v25.OBR{
							SetID:                      "1",
							PlacerOrderNumber:          &v25.EI{EntityIdentifier: "ORD001"},
							UniversalServiceIdentifier: v25.CE{Identifier: "CBC"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "WBC"},
								},
							},
							{
								OBX: &v25.OBX{
									SetID:                 "2",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "RBC"},
								},
							},
							{
								OBX: &v25.OBX{
									SetID:                 "3",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "HGB"},
								},
							},
						},
					},
					{
						ORC: &v25.ORC{
							OrderControl:      "RE",
							PlacerOrderNumber: &v25.EI{EntityIdentifier: "ORD002"},
							FillerOrderNumber: &v25.EI{EntityIdentifier: "FILL002"},
						},
						OBR: &v25.OBR{
							SetID:                      "2",
							PlacerOrderNumber:          &v25.EI{EntityIdentifier: "ORD002"},
							UniversalServiceIdentifier: v25.CE{Identifier: "CMP"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "GLU"},
								},
							},
							{
								OBX: &v25.OBX{
									SetID:                 "2",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "BUN"},
								},
							},
						},
					},
					{
						OBR: &v25.OBR{
							SetID:                      "3",
							UniversalServiceIdentifier: v25.CE{Identifier: "UA"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "ST",
									ObservationIdentifier: v25.CE{Identifier: "UA-COLOR"},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

// verifyRoundTrip encodes the message, decodes it back, and verifies it matches.
func verifyRoundTrip(t *testing.T, original any, registry Registry) {
	t.Helper()

	enc := NewEncoder(nil)
	encoded, err := enc.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Log the encoded message for debugging
	t.Logf("Encoded message:\n%s", bytes.ReplaceAll(encoded, []byte{'\r'}, []byte{'\n'}))

	dec := NewDecoder(registry, nil)
	list, err := dec.DecodeList(encoded)
	if err != nil {
		t.Fatalf("DecodeList failed: %v", err)
	}

	grouped, err := dec.DecodeGroup(list)
	if err != nil {
		t.Fatalf("DecodeGroup failed: %v", err)
	}

	// Re-encode the decoded message to compare
	reencoded, err := enc.Encode(grouped)
	if err != nil {
		t.Fatalf("Re-encode failed: %v", err)
	}

	// The encoded bytes should match
	if !bytes.Equal(encoded, reencoded) {
		t.Logf("Re-encoded message:\n%s", bytes.ReplaceAll(reencoded, []byte{'\r'}, []byte{'\n'}))
		t.Fatalf("Round-trip mismatch:\nOriginal:\n%s\n\nRe-encoded:\n%s",
			bytes.ReplaceAll(encoded, []byte{'\r'}, []byte{'\n'}),
			bytes.ReplaceAll(reencoded, []byte{'\r'}, []byte{'\n'}))
	}

	// Also compare the struct values using reflect.DeepEqual
	if !reflect.DeepEqual(original, grouped) {
		t.Logf("Original type: %T, Grouped type: %T", original, grouped)
		compareStructs(t, "root", reflect.ValueOf(original), reflect.ValueOf(grouped), 0)
		t.Fatalf("Struct comparison failed after round-trip")
	}
}

// compareStructs recursively compares two structs and logs differences
func compareStructs(t *testing.T, path string, orig, grouped reflect.Value, depth int) {
	t.Helper()
	if depth > 10 {
		return // Prevent infinite recursion
	}

	// Handle pointers
	if orig.Kind() == reflect.Pointer {
		if grouped.Kind() != reflect.Pointer {
			t.Logf("  %s: kind mismatch - orig is pointer, grouped is %v", path, grouped.Kind())
			return
		}
		if orig.IsNil() != grouped.IsNil() {
			t.Logf("  %s: nil mismatch - orig nil=%v, grouped nil=%v", path, orig.IsNil(), grouped.IsNil())
			return
		}
		if !orig.IsNil() {
			compareStructs(t, path, orig.Elem(), grouped.Elem(), depth)
		}
		return
	}

	// Handle slices
	if orig.Kind() == reflect.Slice {
		if grouped.Kind() != reflect.Slice {
			t.Logf("  %s: kind mismatch - orig is slice, grouped is %v", path, grouped.Kind())
			return
		}
		if orig.Len() != grouped.Len() {
			t.Logf("  %s: len mismatch - orig=%d, grouped=%d", path, orig.Len(), grouped.Len())
			return
		}
		for i := 0; i < orig.Len(); i++ {
			compareStructs(t, path+"["+string(rune('0'+i))+"]", orig.Index(i), grouped.Index(i), depth+1)
		}
		return
	}

	// Handle structs
	if orig.Kind() == reflect.Struct {
		if grouped.Kind() != reflect.Struct {
			t.Logf("  %s: kind mismatch - orig is struct, grouped is %v", path, grouped.Kind())
			return
		}
		origType := orig.Type()
		for i := 0; i < orig.NumField(); i++ {
			fieldName := origType.Field(i).Name
			origField := orig.Field(i)
			groupedField := grouped.Field(i)
			if !reflect.DeepEqual(origField.Interface(), groupedField.Interface()) {
				compareStructs(t, path+"."+fieldName, origField, groupedField, depth+1)
			}
		}
		return
	}

	// For other types, just log the difference
	if !reflect.DeepEqual(orig.Interface(), grouped.Interface()) {
		t.Logf("  %s: value mismatch - orig=%v, grouped=%v", path, orig.Interface(), grouped.Interface())
	}
}

// TestDecodeGroupEdgeCases tests edge cases in group decoding
func TestDecodeGroupEdgeCases(t *testing.T) {
	t.Run("EmptyOptionalGroups", testEmptyOptionalGroups)
	t.Run("OnlyRequiredFields", testOnlyRequiredFields)
	t.Run("BackwardsSegmentMatch", testBackwardsSegmentMatch)
}

func testEmptyOptionalGroups(t *testing.T) {
	// ORU_R01 with no patient (Patient is optional in PatientResult)
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG006",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				// No Patient - should be valid since Patient is optional
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "TEST"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ValueType:             "NM",
									ObservationIdentifier: v25.CE{Identifier: "VALUE"},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testOnlyRequiredFields(t *testing.T) {
	// Minimal ORL_O34 with only required fields
	msg := v25.ORL_O34{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORL",
				TriggerEvent:     "O34",
				MessageStructure: "ORL_O34",
			},
			MessageControlID: "MSG007",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		MSA: &v25.MSA{
			AcknowledgmentCode: "AA",
			MessageControlID:   "REF002",
		},
		// Response is optional, so we can omit it
	}

	verifyRoundTrip(t, msg, v25.Registry)
}

func testBackwardsSegmentMatch(t *testing.T) {
	// Test case where segments might need backward searching
	// This happens when the same segment type appears in multiple groups
	// and the decoder needs to find the right placement.
	msg := v25.ORU_R01{
		MSH: &v25.MSH{
			FieldSeparator:     "|",
			EncodingCharacters: `^~\&`,
			SendingApplication: &v25.HD{NamespaceID: "LAB"},
			DateTimeOfMessage:  time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			MessageType: v25.MSG{
				MessageCode:      "ORU",
				TriggerEvent:     "R01",
				MessageStructure: "ORU_R01",
			},
			MessageControlID: "MSG008",
			ProcessingID:     v25.PT{ProcessingID: "P"},
			VersionID:        v25.VID{VersionID: "2.5"},
		},
		PatientResult: []v25.ORU_R01_PatientResult{
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Test1", GivenName: "Patient1"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "TEST1"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ObservationIdentifier: v25.CE{Identifier: "OBS1"},
								},
							},
						},
					},
				},
			},
			{
				Patient: &v25.ORU_R01_Patient{
					PID: &v25.PID{
						SetID: "1",
						PatientName: []v25.XPN{
							{FamilyName: "Test2", GivenName: "Patient2"},
						},
					},
				},
				OrderObservation: []v25.ORU_R01_OrderObservation{
					{
						OBR: &v25.OBR{
							SetID:                      "1",
							UniversalServiceIdentifier: v25.CE{Identifier: "TEST2"},
						},
						Observation: []v25.ORU_R01_Observation{
							{
								OBX: &v25.OBX{
									SetID:                 "1",
									ObservationIdentifier: v25.CE{Identifier: "OBS2"},
								},
							},
						},
					},
				},
			},
		},
	}

	verifyRoundTrip(t, msg, v25.Registry)
}
