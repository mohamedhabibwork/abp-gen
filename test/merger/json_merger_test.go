package merger_test

import (
	"encoding/json"
	"testing"

	"github.com/mohamedhabibwork/abp-gen/internal/merger"
)

func TestJSONMerger_Merge(t *testing.T) {
	jsonMerger := merger.NewJSONMerger()

	existing := `{
  "Product": "Product",
  "ProductName": "Product Name"
}`

	newContent := `{
  "Order": "Order",
  "OrderDate": "Order Date"
}`

	merged, conflicts, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	if len(conflicts) > 0 {
		t.Errorf("Expected no conflicts, got %d", len(conflicts))
	}

	// Verify merged JSON contains all keys
	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	if len(mergedData) != 4 {
		t.Errorf("Expected 4 keys in merged JSON, got %d", len(mergedData))
	}
}
