package manager

import (
	"encoding/json"
	"testing"
)

func TestPackageInfoJsonTag(t *testing.T) {
	// Populate all fields to ensure they are present in the JSON output.
	testPackage := PackageInfo{
		Name:           "test-pkg",
		Version:        "1.0.0",
		NewVersion:     "1.1.0",
		Status:         PackageStatusUpgradable,
		Category:       "test-cat",
		Arch:           "amd64",
		PackageManager: "test-pm",
		AdditionalData: map[string]string{"foo": "bar"},
	}

	jsonAsByte, err := json.Marshal(testPackage)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonAsByte, &data); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedTags := []string{"name", "version", "new_version", "status", "category", "arch", "package_manager", "additional_data"}
	for _, tag := range expectedTags {
		if _, ok := data[tag]; !ok {
			t.Errorf("Expected tag '%s' not found in JSON output: %s", tag, string(jsonAsByte))
		}
	}

	if len(data) != len(expectedTags) {
		t.Errorf("Expected %d tags, but got %d. JSON: %s", len(expectedTags), len(data), string(jsonAsByte))
	}
}
