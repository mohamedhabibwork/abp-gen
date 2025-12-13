# Merger Tests

This directory contains unit tests for the smart file merge system.

## Running Tests

```bash
# Run all merger tests
go test ./test/merger/...

# Run with verbose output
go test -v ./test/merger/...

# Run specific test
go test -v ./test/merger/ -run TestClassifier_ClassifyFile
```

## Test Coverage

- **classifier_test.go**: Tests file type classification
- **pattern_merger_test.go**: Tests pattern-based merging for permissions, DbContext
- **json_merger_test.go**: Tests JSON file merging

## Manual Testing

To manually test the merge functionality:

1. Generate code with an existing ABP solution
2. Modify a schema and regenerate
3. The tool will detect existing files and prompt for merge decisions
4. Verify merged files contain both old and new content

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

