package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	// Test that the application can be built
	cmd := exec.Command("go", "build", "-o", "/tmp/aws-switcher-test", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, output)
	}
	t.Log("Application builds successfully")
}

func TestVersion(t *testing.T) {
	// First build the application
	buildCmd := exec.Command("go", "build", "-o", "/tmp/aws-switcher-test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	// Test that the application can show help
	cmd := exec.Command("/tmp/aws-switcher-test", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run application with --help: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "aws-switcher") {
		t.Errorf("Help output should contain 'aws-switcher', got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "AWS SSO") {
		t.Errorf("Help output should contain 'AWS SSO', got: %s", outputStr)
	}

	t.Log("Application runs and shows help successfully")
}
