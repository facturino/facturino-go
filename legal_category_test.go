package facturino

import (
	"encoding/json"
	"reflect"
	"testing"
)

func decodeTaxDecisionBody(t *testing.T, body map[string]interface{}) *TaxDecision {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var decision TaxDecision
	if err := json.Unmarshal(raw, &decision); err != nil {
		t.Fatal(err)
	}
	return &decision
}

func TestLegalCategoryContractResponse(t *testing.T) {
	decision := decodeTaxDecisionBody(t, corpusBody(t, "legal-category"))
	category := decision.LegalCategory
	if category == nil {
		t.Fatal("LegalCategory = nil")
	}
	if category.Source != "sirene" || category.SIREN != "123456789" ||
		category.Status != "found" || category.LegalCategoryCode == nil ||
		*category.LegalCategoryCode != "5710" || category.CheckedAt != "2026-09-15" {
		t.Fatalf("LegalCategory = %+v, want the complete SIRENE consultation", category)
	}

	// The SDK-only fixture extends the shared response without changing it.
	body := corpusBody(t, "legal-category")
	delete(body, "legalCategory")
	if !reflect.DeepEqual(body, corpusBody(t, "taxDecision")) {
		t.Error("legal-category fixture must be taxDecision plus legalCategory only")
	}
}

func TestLegalCategoryHistoricalOmission(t *testing.T) {
	decision := decodeTaxDecisionBody(t, corpusBody(t, "taxDecision"))
	if decision.LegalCategory != nil {
		t.Errorf("historical LegalCategory = %+v, want nil", decision.LegalCategory)
	}
	raw, err := json.Marshal(decision)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, exists := fields["legalCategory"]; exists {
		t.Error("serialization invented a historical legalCategory")
	}
}

func TestLegalCategoryNullableCodeAndOpenStrings(t *testing.T) {
	for _, status := range []string{"not_found", "unavailable", "future_status"} {
		t.Run(status, func(t *testing.T) {
			body := corpusBody(t, "legal-category")
			category := body["legalCategory"].(map[string]interface{})
			category["status"] = status
			category["legalCategoryCode"] = nil
			body["status"] = "pending_verification"
			body["totals"] = nil
			body["amountToCharge"] = nil
			for _, line := range body["lines"].([]interface{}) {
				for _, key := range []string{"amountHT", "amountVAT", "amountTTC", "rateCentipercent"} {
					line.(map[string]interface{})[key] = nil
				}
			}
			body["issues"] = []interface{}{map[string]interface{}{
				"code": "business_status_not_evidenced", "detailCode": "legal_category_unresolved",
				"message": "SIRENE could not give the legal category; complete the buyer status evidence.",
			}}
			decision := decodeTaxDecisionBody(t, body)
			if decision.LegalCategory == nil || decision.LegalCategory.Status != status ||
				decision.LegalCategory.LegalCategoryCode != nil {
				t.Fatalf("LegalCategory = %+v, want status %q and nil code", decision.LegalCategory, status)
			}
			if decision.IsFinal() || decision.Totals != nil || decision.AmountToCharge != nil ||
				len(decision.Issues) != 1 || decision.Issues[0].DetailCode != "legal_category_unresolved" {
				t.Fatalf("lost the pending decision or its issue: %+v", decision)
			}
			raw, err := json.Marshal(decision.LegalCategory)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["legalCategoryCode"]) != "null" {
				t.Errorf("legalCategoryCode = %s, want explicit null", fields["legalCategoryCode"])
			}
		})
	}

	body := corpusBody(t, "legal-category")
	body["legalCategory"].(map[string]interface{})["legalCategoryCode"] = "future_category"
	decision := decodeTaxDecisionBody(t, body)
	if decision.LegalCategory == nil || decision.LegalCategory.LegalCategoryCode == nil ||
		*decision.LegalCategory.LegalCategoryCode != "future_category" {
		t.Fatalf("LegalCategory = %+v, want an open category code", decision.LegalCategory)
	}
}
