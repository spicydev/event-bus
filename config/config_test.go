package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const testConfigContent = `
spanner:
  project_id: test-project
  instance_id: test-instance
  database_id: test-database
pubsub:
  project_id: test-project
  subscriptions:
    - id: "sub1"
      topic_id: "topic1"
      target_table: "table1"
      entity_type: "type1"
    - id: "sub2"
      topic_id: "topic2"
      target_table: "table2"
      entity_type: "type2"
server:
  shutdown_delay: 5
`

func TestLoadConfigWithMultipleSubscriptions(t *testing.T) {
	// Create a temporary directory for the test config file
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Write the test config file
	configFilePath := filepath.Join(tmpDir, "config-test.yaml")
	err = os.WriteFile(configFilePath, []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Set GO_PROFILE to "test" and original working directory
	// The LoadConfig function expects the config file to be in the current working directory
	// or a directory specified by an absolute path if we were to change LoadConfig.
	// For simplicity, we change the CWD for this test.
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change current working directory to temp dir: %v", err)
	}
	defer os.Chdir(originalWd) // Restore original working directory

	originalGoProfile := os.Getenv("GO_PROFILE")
	os.Setenv("GO_PROFILE", "test")
	defer os.Setenv("GO_PROFILE", originalGoProfile) // Restore original GO_PROFILE

	// Call LoadConfig
	props, err := LoadConfig()

	// Assertions
	assert.NoError(t, err, "LoadConfig() should not return an error")
	assert.NotNil(t, props, "LoadConfig() should return a non-nil Props struct")

	assert.NotNil(t, props.PubSub, "Props.PubSub should not be nil")
	assert.Len(t, props.PubSub.Subscriptions, 2, "Should load 2 subscriptions")

	// Subscription 1
	assert.Equal(t, "sub1", props.PubSub.Subscriptions[0].ID, "Subscription 1 ID mismatch")
	assert.Equal(t, "topic1", props.PubSub.Subscriptions[0].TopicID, "Subscription 1 TopicID mismatch")
	assert.Equal(t, "table1", props.PubSub.Subscriptions[0].TargetTable, "Subscription 1 TargetTable mismatch")
	assert.Equal(t, "type1", props.PubSub.Subscriptions[0].EntityType, "Subscription 1 EntityType mismatch")

	// Subscription 2
	assert.Equal(t, "sub2", props.PubSub.Subscriptions[1].ID, "Subscription 2 ID mismatch")
	assert.Equal(t, "topic2", props.PubSub.Subscriptions[1].TopicID, "Subscription 2 TopicID mismatch")
	assert.Equal(t, "table2", props.PubSub.Subscriptions[1].TargetTable, "Subscription 2 TargetTable mismatch")
	assert.Equal(t, "type2", props.PubSub.Subscriptions[1].EntityType, "Subscription 2 EntityType mismatch")

	assert.Equal(t, "test-project", props.Spanner.ProjectID, "Spanner ProjectID mismatch")
	assert.Equal(t, int32(5), props.Server.ShutdownDelay, "Server ShutdownDelay mismatch")
}

// TestLoadConfigNoProfile tests if config.yaml is loaded when GO_PROFILE is not set.
func TestLoadConfigNoProfile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-no-profile-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy config.yaml
	dummyConfigContent := `
spanner:
  project_id: default-project
pubsub:
  project_id: default-project
  subscriptions:
    - id: "default-sub"
      topic_id: "default-topic"
      target_table: "default-table"
      entity_type: "default-type"
server:
  shutdown_delay: 1
`
	configFilePath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFilePath, []byte(dummyConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config.yaml file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change current working directory to temp dir: %v", err)
	}
	defer os.Chdir(originalWd)

	originalGoProfile := os.Getenv("GO_PROFILE")
	os.Unsetenv("GO_PROFILE") // Ensure GO_PROFILE is not set
	defer os.Setenv("GO_PROFILE", originalGoProfile)

	props, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "default-project", props.Spanner.ProjectID)
	assert.Len(t, props.PubSub.Subscriptions, 1)
	assert.Equal(t, "default-sub", props.PubSub.Subscriptions[0].ID)
}

// TestLoadConfigNonExistentProfile tests behavior when GO_PROFILE is set but file doesn't exist.
func TestLoadConfigNonExistentProfile(t *testing.T) {
	originalWd, err := os.Getwd() // Assuming config.yaml might exist in original WD
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a temporary directory that will be empty
	tmpDir, err := os.MkdirTemp("", "config-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change current working directory to temp dir: %v", err)
	}
	defer os.Chdir(originalWd)


	originalGoProfile := os.Getenv("GO_PROFILE")
	os.Setenv("GO_PROFILE", "nonexistentprofile")
	defer os.Setenv("GO_PROFILE", originalGoProfile)

	_, err = LoadConfig()
	assert.Error(t, err, "LoadConfig should return an error for a non-existent profile config file")
	// Check if the error is due to file not existing
	assert.True(t, os.IsNotExist(err), "Error should be os.IsNotExist")
}

// TestLoadConfigInvalidYaml tests behavior with a malformed YAML file.
func TestLoadConfigInvalidYaml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-invalid-yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidYamlContent := `
spanner:
  project_id: test-project
pubsub:
  project_id: test-project
  subscriptions:
    - id: "sub1"
      topic_id: "topic1"
      target_table: "table1"
      entity_type: "type1"
  server: # Incorrect indentation
    shutdown_delay: 5
`
	configFilePath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFilePath, []byte(invalidYamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config.yaml file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change current working directory to temp dir: %v", err)
	}
	defer os.Chdir(originalWd)

	originalGoProfile := os.Getenv("GO_PROFILE")
	os.Unsetenv("GO_PROFILE") // Use config.yaml
	defer os.Setenv("GO_PROFILE", originalGoProfile)

	_, err = LoadConfig()
	assert.Error(t, err, "LoadConfig should return an error for invalid YAML")
	// Optionally, check for a specific YAML error type if the library provides one
	// For now, just asserting that an error of type *yaml.TypeError occurs
	_, isYamlTypeError := err.(*yaml.TypeError)
	assert.True(t, isYamlTypeError, "Error should be a yaml.TypeError for malformed content")
}
