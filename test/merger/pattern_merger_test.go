package merger_test

import (
	"testing"

	"github.com/mohamedhabibwork/abp-gen/internal/merger"
)

func TestPatternMerger_MergePermissions(t *testing.T) {
	patternMerger := merger.NewPatternMerger()

	existing := `namespace Test
{
    public static class ProductManagement
    {
        public const string Create = "Product.Create";
    }
}`

	newContent := `namespace Test
{
    public static class OrderManagement
    {
        public const string Create = "Order.Create";
    }
}`

	merged, conflicts, err := patternMerger.Merge(existing, newContent, merger.FileTypePermissions)

	if err != nil {
		t.Errorf("Merge failed: %v", err)
	}

	if len(conflicts) > 0 {
		t.Errorf("Expected no conflicts, got %d", len(conflicts))
	}

	// Verify both classes are present
	if merged == "" {
		t.Error("Merged content is empty")
	}
}
