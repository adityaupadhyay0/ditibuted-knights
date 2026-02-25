package simulation

import (
	"testing"

	"github.com/user/infratwin/backend/graph"
	"github.com/user/infratwin/backend/parser"
)

func TestRunSimulation(t *testing.T) {
	config := &parser.InfraConfig{
		Resources: []parser.Resource{
			{Type: "aws_vpc", Name: "main"},
			{Type: "aws_subnet", Name: "sub", Dependencies: []string{"aws_vpc.main"}},
		},
	}
	g := graph.BuildGraph(config)

	params := &SimulationParams{
		DurationSeconds: 5,
		Traffic: []TrafficPattern{
			{NodeID: "aws_subnet.sub", RequestsPerSec: 200},
		},
	}

	result := RunSimulation(config, g, params)

	if len(result.Timeline) != 5 {
		t.Errorf("expected 5 timestamps, got %d", len(result.Timeline))
	}

	// Check if traffic propagated to VPC
	lastTick := result.Timeline[4]
	vpcMetrics, ok := lastTick.Nodes["aws_vpc.main"]
	if !ok {
		t.Fatal("vpc metrics not found in simulation result")
	}

	if vpcMetrics.Requests != 200 {
		t.Errorf("expected 200 RPS on VPC due to propagation, got %f", vpcMetrics.Requests)
	}

	if vpcMetrics.Replicas <= 1 {
		t.Errorf("expected VPC to scale up from 200 RPS, got %d replicas", vpcMetrics.Replicas)
	}
}
