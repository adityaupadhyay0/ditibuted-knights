package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Resource struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

type Module struct {
	Name         string   `json:"name"`
	Source       string   `json:"source"`
	Dependencies []string `json:"dependencies"`
}

type InfraConfig struct {
	Resources []Resource `json:"resources"`
	Modules   []Module   `json:"modules"`
}

func ParseTerraform(dir string) (*InfraConfig, error) {
	parser := hclparse.NewParser()
	config := &InfraConfig{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			file, diags := parser.ParseHCLFile(path)
			if diags.HasErrors() {
				return diags
			}

			resources, modules, err := processFile(file)
			if err != nil {
				return err
			}
			config.Resources = append(config.Resources, resources...)
			config.Modules = append(config.Modules, modules...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}

func processFile(file *hcl.File) ([]Resource, []Module, error) {
	var resources []Resource
	var modules []Module

	content, _, diags := file.Body.PartialContent(terraformSchema)
	if diags.HasErrors() {
		// We can ignore some errors if it's just unknown blocks for MVP
	}

	for _, block := range content.Blocks {
		if block.Type == "resource" {
			res := Resource{
				Type: block.Labels[0],
				Name: block.Labels[1],
			}
			res.Dependencies = extractDependencies(block.Body)
			resources = append(resources, res)
		} else if block.Type == "module" {
			mod := Module{
				Name: block.Labels[0],
			}
			// Extract source from module body
			attrs, _ := block.Body.JustAttributes()
			if attr, ok := attrs["source"]; ok {
				val, _ := attr.Expr.Value(nil)
				mod.Source = val.AsString()
			}
			mod.Dependencies = extractDependencies(block.Body)
			modules = append(modules, mod)
		}
	}

	return resources, modules, nil
}

var terraformSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type: "variable",
			LabelNames: []string{"name"},
		},
		{
			Type: "output",
			LabelNames: []string{"name"},
		},
		{
			Type: "provider",
			LabelNames: []string{"name"},
		},
		{
			Type: "terraform",
		},
        {
            Type: "data",
            LabelNames: []string{"type", "name"},
        },
	},
}

func extractDependencies(body hcl.Body) []string {
	var deps []string
	attrs, _ := body.JustAttributes()
	for _, attr := range attrs {
		for _, traversal := range attr.Expr.Variables() {
			if len(traversal) >= 2 {
				// Simple dependency detection: resource_type.resource_name
                // or module.module_name
				dep := fmt.Sprintf("%s.%s", traversal[0].(hcl.TraverseRoot).Name, traversal[1].(hcl.TraverseAttr).Name)
				deps = append(deps, dep)
			}
		}
	}
	return deps
}
