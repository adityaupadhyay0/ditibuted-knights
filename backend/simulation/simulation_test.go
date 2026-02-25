package simulation

import (
	"testing"

	"github.com/user/distributed-knights/backend/graph"
	"github.com/user/distributed-knights/backend/parser"
)

func TestRunCampaign(t *testing.T) {
	config := &parser.InfraConfig{
		Resources: []parser.Resource{
			{Type: "aws_vpc", Name: "main"},
			{Type: "aws_subnet", Name: "sub", Dependencies: []string{"aws_vpc.main"}},
		},
	}
	g := graph.BuildGraph(config)

	params := &CampaignParams{
		DurationSeconds: 5,
		Traffic: []MessengerPattern{
			{NodeID: "aws_subnet.sub", RequestsPerSec: 200},
		},
	}

	result := RunCampaign(config, g, params)

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

func TestCampaignFailures(t *testing.T) {
	config := &parser.InfraConfig{
		Resources: []parser.Resource{
			{Type: "aws_vpc", Name: "main"},
			{Type: "aws_subnet", Name: "sub", Dependencies: []string{"aws_vpc.main"}},
		},
	}
	g := graph.BuildGraph(config)

	params := &CampaignParams{
		DurationSeconds: 10,
		Traffic: []MessengerPattern{
			{NodeID: "aws_subnet.sub", RequestsPerSec: 100},
		},
		Failures: []SiegeEvent{
			{
				Type: SiegeDragonStrike,
				NodeID: "aws_vpc.main",
				StartTime: 3,
				EndTime: 6,
			},
		},
	}

	result := RunCampaign(config, g, params)

	// At T=0, VPC should be UP
	if result.Timeline[0].Nodes["aws_vpc.main"].Status != StatusStanding {
		t.Errorf("expected VPC to be UP at T=0")
	}

	// At T=4, VPC should be DOWN
	if result.Timeline[4].Nodes["aws_vpc.main"].Status != StatusFallen {
		t.Errorf("expected VPC to be DOWN at T=4")
	}

	// At T=4, Subnet should be DEGRADED because its dependency (VPC) is down
	if result.Timeline[4].Nodes["aws_subnet.sub"].Status != StatusBesieged {
		t.Errorf("expected Subnet to be DEGRADED at T=4, got %s", result.Timeline[4].Nodes["aws_subnet.sub"].Status)
	}

	if result.Timeline[4].Nodes["aws_subnet.sub"].ErrorRate < 50 {
		t.Errorf("expected Subnet to have high error rate at T=4, got %f", result.Timeline[4].Nodes["aws_subnet.sub"].ErrorRate)
	}

	// At T=8, VPC should be UP again
	if result.Timeline[8].Nodes["aws_vpc.main"].Status != StatusStanding {
		t.Errorf("expected VPC to be UP at T=8")
	}
}
