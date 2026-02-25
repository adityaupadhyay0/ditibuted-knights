package graph

import (
	"github.com/user/infratwin/backend/parser"
)

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func BuildGraph(config *parser.InfraConfig) *Graph {
	g := &Graph{
		Nodes: make([]Node, 0),
		Edges: make([]Edge, 0),
	}

	nodeMap := make(map[string]bool)

	// Add Resources
	for _, res := range config.Resources {
		id := res.Type + "." + res.Name
		g.Nodes = append(g.Nodes, Node{
			ID:    id,
			Label: id,
			Type:  "resource",
		})
		nodeMap[id] = true
	}

	// Add Modules
	for _, mod := range config.Modules {
		id := "module." + mod.Name
		g.Nodes = append(g.Nodes, Node{
			ID:    id,
			Label: id,
			Type:  "module",
		})
		nodeMap[id] = true
	}

	// Add Edges
	for _, res := range config.Resources {
		target := res.Type + "." + res.Name
		for _, dep := range res.Dependencies {
			if nodeMap[dep] {
				g.Edges = append(g.Edges, Edge{
					ID:     dep + "->" + target,
					Source: dep,
					Target: target,
				})
			}
		}
	}

	for _, mod := range config.Modules {
		target := "module." + mod.Name
		for _, dep := range mod.Dependencies {
			if nodeMap[dep] {
				g.Edges = append(g.Edges, Edge{
					ID:     dep + "->" + target,
					Source: dep,
					Target: target,
				})
			}
		}
	}

	return g
}
