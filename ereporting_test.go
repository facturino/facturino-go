package facturino

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEReportingLineParamsVATExCodeMarshalling(t *testing.T) {
	base := EReportingLineParams{Category: "international", Amount: 10000, VATRate: 0, VATAmount: 0}

	omitted, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(omitted), "vatexCode") {
		t.Errorf("nil VATExCode must omit the field, got %s", omitted)
	}

	code := "VATEX-EU-G"
	valued := base
	valued.VATExCode = &code
	got, err := json.Marshal(valued)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"vatexCode":"VATEX-EU-G"`) {
		t.Errorf("valued VATExCode must be sent, got %s", got)
	}

	explicit := base
	explicit.VATExCodeNull = true
	got, err = json.Marshal(explicit)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"vatexCode":null`) {
		t.Errorf("VATExCodeNull must send an explicit null, got %s", got)
	}
	if strings.Contains(string(got), "VATExCodeNull") {
		t.Errorf("the flag itself must not travel, got %s", got)
	}
	if !strings.Contains(string(got), `"category":"international"`) {
		t.Errorf("other fields must still be sent, got %s", got)
	}
}
