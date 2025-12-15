# Merger Tests

This directory contains comprehensive unit tests for the smart file merge system.

## Running Tests

```bash
# Run all merger tests
go test ./test/merger/...

# Run with verbose output
go test -v ./test/merger/...

# Run specific test
go test -v ./test/merger/ -run TestJSONMerger_Merge

# Run with coverage
go test -cover ./test/merger/...

# Generate coverage report
go test -coverprofile=coverage.out ./test/merger/...
go tool cover -html=coverage.out
```

## Test Coverage

### classifier_test.go
Tests file type classification and merge strategy selection:
- File type detection (Permissions, Entity, DTO, Service, etc.)
- Merge strategy mapping (Pattern, AST, JSON, None)

### pattern_merger_test.go
Tests pattern-based merging for C# code files:
- Permissions file merging
- DbContext merging
- Class-level pattern recognition

### json_merger_test.go
Comprehensive tests for JSON file merging with conflict strategies:
- **Basic merging**: Simple key-value merging without conflicts
- **Conflict detection**: Detecting and reporting conflicts
- **Strategy: Overwrite**: Overwriting existing values with new ones
- **Strategy: Skip**: Keeping existing values, skipping new ones
- **Strategy: Append**: Intelligent merging with conflict detection (default)
- **Nested objects**: Recursive merging of nested JSON structures
- **Array merging**: Merging arrays with duplicate removal
- **Localization merging**: Specific tests for localization JSON files
- **Edge cases**: Empty inputs, invalid JSON, etc.

## Test Scenarios

### Scenario 1: Add New Entity
- Generate Product entity
- Add Order entity to schema
- Regenerate
- Verify permissions file contains both Product and Order permissions

### Scenario 2: Merge Entity Properties
- Generate Product entity with Name property
- Add Price property to schema
- Regenerate with merge mode
- Verify Product entity contains both properties

### Scenario 3: JSON Localization Merge
- Generate with English localization
- Add Arabic localization
- Regenerate
- Verify localization JSON contains both languages

### Scenario 4: Conflict Resolution Strategies
- Test overwrite strategy: New values replace existing
- Test skip strategy: Existing values are preserved
- Test append strategy: Conflicts are detected and reported

## JSON Merger Test Details

### Conflict Strategies Tested

1. **Overwrite Strategy** (`--merge-strategy=overwrite`)
   - New values replace existing values
   - No conflicts reported
   - Useful for forced updates

2. **Skip Strategy** (`--merge-strategy=skip`)
   - Existing values are preserved
   - New values are ignored for existing keys
   - New keys are still added

3. **Append Strategy** (default)
   - Intelligent merging of nested objects
   - Array merging with duplicate removal
   - Conflicts detected and reported for different values

### Test Cases

- ✅ Basic merge without conflicts
- ✅ Conflict detection and reporting
- ✅ Overwrite strategy behavior
- ✅ Skip strategy behavior
- ✅ Append strategy behavior
- ✅ Nested object merging
- ✅ Array merging with duplicates
- ✅ Localization-specific scenarios
- ✅ Empty input handling
- ✅ Invalid JSON error handling

## Manual Testing

To manually test the merge functionality:

1. Generate code with an existing ABP solution
2. Modify a schema and regenerate
3. The tool will detect existing files and prompt for merge decisions
4. Verify merged files contain both old and new content

## Example Test Output

```bash
$ go test -v ./test/merger/...

=== RUN   TestJSONMerger_Merge
--- PASS: TestJSONMerger_Merge (0.00s)
=== RUN   TestJSONMerger_MergeWithConflicts
--- PASS: TestJSONMerger_MergeWithConflicts (0.00s)
=== RUN   TestJSONMerger_StrategyOverwrite
--- PASS: TestJSONMerger_StrategyOverwrite (0.00s)
=== RUN   TestJSONMerger_StrategySkip
--- PASS: TestJSONMerger_StrategySkip (0.00s)
=== RUN   TestJSONMerger_StrategyAppend
--- PASS: TestJSONMerger_StrategyAppend (0.00s)
=== RUN   TestJSONMerger_NestedObjects
--- PASS: TestJSONMerger_NestedObjects (0.00s)
=== RUN   TestJSONMerger_MergeArrays
--- PASS: TestJSONMerger_MergeArrays (0.00s)
=== RUN   TestJSONMerger_LocalizationMerge
--- PASS: TestJSONMerger_LocalizationMerge (0.00s)
=== RUN   TestJSONMerger_EmptyInputs
--- PASS: TestJSONMerger_EmptyInputs (0.00s)
PASS
ok      github.com/mohamedhabibwork/abp-gen/test/merger    0.123s
```

