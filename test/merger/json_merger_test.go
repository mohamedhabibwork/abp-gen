package merger_test

import (
	"encoding/json"
	"testing"

	"github.com/mohamedhabibwork/abp-gen/internal/merger"
)

// TestJSONMerger_Merge tests basic JSON merging without conflicts
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

	// Verify all keys are present
	expectedKeys := []string{"Product", "ProductName", "Order", "OrderDate"}
	for _, key := range expectedKeys {
		if _, exists := mergedData[key]; !exists {
			t.Errorf("Expected key '%s' not found in merged JSON", key)
		}
	}
}

// TestJSONMerger_MergeWithConflicts tests conflict detection
func TestJSONMerger_MergeWithConflicts(t *testing.T) {
	jsonMerger := merger.NewJSONMerger()

	existing := `{
  "Product": "Product",
  "ProductName": "Product Name"
}`

	newContent := `{
  "Product": "Different Product",
  "Order": "Order"
}`

	merged, conflicts, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	// Should detect conflict for "Product" key
	if len(conflicts) == 0 {
		t.Error("Expected conflicts, got none")
	}

	if len(conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].Type != merger.ConflictTypeDifferentValue {
		t.Errorf("Expected conflict type %v, got %v", merger.ConflictTypeDifferentValue, conflicts[0].Type)
	}

	// Merged should return existing content when conflicts exist
	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// Should contain original "Product" value
	if mergedData["Product"] != "Product" {
		t.Errorf("Expected 'Product' value to remain unchanged, got %v", mergedData["Product"])
	}
}

// TestJSONMerger_StrategyOverwrite tests overwrite strategy
func TestJSONMerger_StrategyOverwrite(t *testing.T) {
	jsonMerger := merger.NewJSONMergerWithStrategy("overwrite")

	existing := `{
  "Product": "Product",
  "ProductName": "Product Name",
  "Order": "Order"
}`

	newContent := `{
  "Product": "New Product",
  "OrderDate": "Order Date"
}`

	merged, _, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	// Overwrite strategy should not report conflicts, it just overwrites
	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// "Product" should be overwritten
	if mergedData["Product"] != "New Product" {
		t.Errorf("Expected 'Product' to be overwritten to 'New Product', got %v", mergedData["Product"])
	}

	// "ProductName" should remain (not in new content)
	if mergedData["ProductName"] != "Product Name" {
		t.Errorf("Expected 'ProductName' to remain, got %v", mergedData["ProductName"])
	}

	// "Order" should remain (not in new content)
	if mergedData["Order"] != "Order" {
		t.Errorf("Expected 'Order' to remain, got %v", mergedData["Order"])
	}

	// "OrderDate" should be added
	if mergedData["OrderDate"] != "Order Date" {
		t.Errorf("Expected 'OrderDate' to be added, got %v", mergedData["OrderDate"])
	}
}

// TestJSONMerger_StrategySkip tests skip strategy
func TestJSONMerger_StrategySkip(t *testing.T) {
	jsonMerger := merger.NewJSONMergerWithStrategy("skip")

	existing := `{
  "Product": "Product",
  "ProductName": "Product Name"
}`

	newContent := `{
  "Product": "New Product",
  "Order": "Order"
}`

	merged, _, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// "Product" should keep existing value (skipped)
	if mergedData["Product"] != "Product" {
		t.Errorf("Expected 'Product' to keep existing value 'Product', got %v", mergedData["Product"])
	}

	// "ProductName" should remain
	if mergedData["ProductName"] != "Product Name" {
		t.Errorf("Expected 'ProductName' to remain, got %v", mergedData["ProductName"])
	}

	// "Order" should be added (new key)
	if mergedData["Order"] != "Order" {
		t.Errorf("Expected 'Order' to be added, got %v", mergedData["Order"])
	}
}

// TestJSONMerger_StrategyAppend tests append strategy (default)
func TestJSONMerger_StrategyAppend(t *testing.T) {
	jsonMerger := merger.NewJSONMergerWithStrategy("append")

	existing := `{
  "Product": "Product",
  "ProductName": "Product Name"
}`

	newContent := `{
  "Product": "New Product",
  "Order": "Order"
}`

	merged, conflicts, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	// Append strategy should detect conflicts for different values
	if len(conflicts) == 0 {
		t.Error("Expected conflicts with append strategy when values differ")
	}

	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// "Product" should keep existing value (conflict detected, so keep existing)
	if mergedData["Product"] != "Product" {
		t.Errorf("Expected 'Product' to keep existing value 'Product', got %v", mergedData["Product"])
	}

	// "Order" should be added (new key)
	if mergedData["Order"] != "Order" {
		t.Errorf("Expected 'Order' to be added, got %v", mergedData["Order"])
	}
}

// TestJSONMerger_NestedObjects tests nested object merging
func TestJSONMerger_NestedObjects(t *testing.T) {
	jsonMerger := merger.NewJSONMerger()

	existing := `{
  "Product": {
    "Name": "Product",
    "Description": "A product"
  },
  "Order": "Order"
}`

	newContent := `{
  "Product": {
    "Price": "Price",
    "Stock": "Stock"
  },
  "Category": "Category"
}`

	merged, _, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// Product should be a nested object
	product, ok := mergedData["Product"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'Product' to be a nested object")
	}

	// Should contain all nested keys
	expectedProductKeys := []string{"Name", "Description", "Price", "Stock"}
	for _, key := range expectedProductKeys {
		if _, exists := product[key]; !exists {
			t.Errorf("Expected nested key '%s' not found in Product", key)
		}
	}

	// Top-level keys should be merged
	if mergedData["Order"] != "Order" {
		t.Errorf("Expected 'Order' to be present, got %v", mergedData["Order"])
	}

	if mergedData["Category"] != "Category" {
		t.Errorf("Expected 'Category' to be present, got %v", mergedData["Category"])
	}
}

// TestJSONMerger_MergeArrays tests array merging
func TestJSONMerger_MergeArrays(t *testing.T) {
	jsonMerger := merger.NewJSONMerger()

	existing := `{
  "Items": ["Item1", "Item2"],
  "Tags": ["Tag1"]
}`

	newContent := `{
  "Items": ["Item3"],
  "Tags": ["Tag1", "Tag2"]
}`

	merged, _, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// Items array should be merged (duplicates removed)
	items, ok := mergedData["Items"].([]interface{})
	if !ok {
		t.Fatal("Expected 'Items' to be an array")
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items in merged array, got %d", len(items))
	}

	// Tags array should have Tag1 only once (duplicate removed)
	tags, ok := mergedData["Tags"].([]interface{})
	if !ok {
		t.Fatal("Expected 'Tags' to be an array")
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags in merged array, got %d", len(tags))
	}
}

// TestJSONMerger_LocalizationMerge tests localization-specific merging
func TestJSONMerger_LocalizationMerge(t *testing.T) {
	jsonMerger := merger.NewJSONMergerWithStrategy("append")

	existing := `{
  "Product": "Product",
  "Product.Name": "Product Name",
  "Product.Description": "Product Description",
  "Permission:Product": "Product Management",
  "Permission:Product.Create": "Create Product"
}`

	newContent := `{
  "Order": "Order",
  "Order.Name": "Order Name",
  "Permission:Order": "Order Management",
  "Permission:Order.Create": "Create Order"
}`

	merged, conflicts, err := jsonMerger.Merge(existing, newContent)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	if len(conflicts) > 0 {
		t.Errorf("Expected no conflicts for localization merge, got %d", len(conflicts))
	}

	var mergedData map[string]interface{}
	if err := json.Unmarshal([]byte(merged), &mergedData); err != nil {
		t.Errorf("Failed to parse merged JSON: %v", err)
	}

	// Verify all keys are present
	expectedKeys := []string{
		"Product", "Product.Name", "Product.Description",
		"Permission:Product", "Permission:Product.Create",
		"Order", "Order.Name",
		"Permission:Order", "Permission:Order.Create",
	}

	for _, key := range expectedKeys {
		if _, exists := mergedData[key]; !exists {
			t.Errorf("Expected localization key '%s' not found", key)
		}
	}
}

// TestJSONMerger_EmptyInputs tests edge cases with empty inputs
func TestJSONMerger_EmptyInputs(t *testing.T) {
	jsonMerger := merger.NewJSONMerger()

	tests := []struct {
		name        string
		existing    string
		newContent  string
		expectError bool
	}{
		{
			name:        "Empty existing",
			existing:    `{}`,
			newContent:  `{"Key": "Value"}`,
			expectError: false,
		},
		{
			name:        "Empty new content",
			existing:    `{"Key": "Value"}`,
			newContent:  `{}`,
			expectError: false,
		},
		{
			name:        "Both empty",
			existing:    `{}`,
			newContent:  `{}`,
			expectError: false,
		},
		{
			name:        "Invalid existing JSON",
			existing:    `{invalid}`,
			newContent:  `{"Key": "Value"}`,
			expectError: true,
		},
		{
			name:        "Invalid new JSON",
			existing:    `{"Key": "Value"}`,
			newContent:  `{invalid}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := jsonMerger.Merge(tt.existing, tt.newContent)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
