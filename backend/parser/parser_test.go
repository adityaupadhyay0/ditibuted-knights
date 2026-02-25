package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTerraform(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-tf-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tfCode := `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfCode), 0644)
	if err != nil {
		t.Fatal(err)
	}

	config, err := ParseTerraform(tmpDir)
	if err != nil {
		t.Fatalf("ParseTerraform failed: %v", err)
	}

	if len(config.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(config.Resources))
	}

	foundSubnet := false
	for _, res := range config.Resources {
		if res.Type == "aws_subnet" {
			foundSubnet = true
			if len(res.Dependencies) == 0 {
				t.Errorf("expected dependencies for aws_subnet, got none")
			}
			foundDep := false
			for _, dep := range res.Dependencies {
				if dep == "aws_vpc.main" {
					foundDep = true
					break
				}
			}
			if !foundDep {
				t.Errorf("expected dependency aws_vpc.main not found in %v", res.Dependencies)
			}
		}
	}
	if !foundSubnet {
		t.Error("aws_subnet not found in parsed resources")
	}
}
