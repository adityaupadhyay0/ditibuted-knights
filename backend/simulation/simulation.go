package simulation

import (
	"math"

	"github.com/user/infratwin/backend/graph"
	"github.com/user/infratwin/backend/parser"
)

type FailureType string

const (
	FailureNodeOutage      FailureType = "node_outage"
	FailureNetworkPartition FailureType = "network_partition"
)

type FailureEvent struct {
	Type      FailureType `json:"type"`
	NodeID    string      `json:"nodeId"`
	TargetID  string      `json:"targetId,omitempty"` // For network partitions
	StartTime int         `json:"startTime"`
	EndTime   int         `json:"endTime"`
}

type TrafficPattern struct {
	NodeID         string  `json:"nodeId"`
	RequestsPerSec float64 `json:"rps"`
}

type SimulationParams struct {
	DurationSeconds int              `json:"durationSeconds"`
	Traffic         []TrafficPattern `json:"traffic"`
	Failures        []FailureEvent   `json:"failures"`
}

type SimulationResult struct {
	Timeline []TimestampResult `json:"timeline"`
}

type TimestampResult struct {
	Timestamp int                     `json:"timestamp"`
	Nodes     map[string]NodeMetrics `json:"nodes"`
}

type NodeStatus string

const (
	StatusUp       NodeStatus = "up"
	StatusDown     NodeStatus = "down"
	StatusDegraded NodeStatus = "degraded"
)

type NodeMetrics struct {
	CPUUsage    float64    `json:"cpuUsage"`    // Percentage 0-100
	MemoryUsage float64    `json:"memoryUsage"` // Percentage 0-100
	Latency     float64    `json:"latency"`     // Milliseconds
	Replicas    int        `json:"replicas"`
	Requests    float64    `json:"requests"`    // Requests handled at this node
	ErrorRate   float64    `json:"errorRate"`   // Percentage 0-100
	Status      NodeStatus `json:"status"`
}

func RunSimulation(config *parser.InfraConfig, g *graph.Graph, params *SimulationParams) *SimulationResult {
	if params.DurationSeconds <= 0 {
		params.DurationSeconds = 10
	}

	result := &SimulationResult{
		Timeline: make([]TimestampResult, params.DurationSeconds),
	}

	// Initialize state
	nodeState := make(map[string]*NodeMetrics)
	for _, node := range g.Nodes {
		nodeState[node.ID] = &NodeMetrics{
			Replicas: 1, // Default to 1 replica
		}
	}

	// Build adjacency list for propagation (reverse of dependencies)
	// If A depends on B, traffic flows from A to B?
	// Usually in cloud, dependencies are "downstream".
	// E.g. Web server depends on DB. Traffic: Web -> DB.
	downstream := make(map[string][]string)
	for _, edge := range g.Edges {
		downstream[edge.Target] = append(downstream[edge.Target], edge.Source)
	}
	// Wait, our edges were Source -> Target where Source is Dependency.
	// So Target depends on Source.
	// Traffic flows from the one that HAS the dependency to the dependency.
	// Example: aws_instance.web depends on aws_subnet.public. Traffic flows web -> subnet? No.
	// Actually, traffic flow is usually inverse of infra dependency in some cases,
	// but let's assume traffic is injected at "entry" nodes and flows to their dependencies.

	for t := 0; t < params.DurationSeconds; t++ {
		currentMetrics := make(map[string]NodeMetrics)

		// Reset requests for this timestep
		for id := range nodeState {
			nodeState[id].Requests = 0
		}

		// 0. Reset status and error rate
		for id := range nodeState {
			nodeState[id].Status = StatusUp
			nodeState[id].ErrorRate = 0
		}

		// 0.1 Apply active failures
		activeFailures := make(map[string]bool)
		for _, f := range params.Failures {
			if t >= f.StartTime && t <= f.EndTime {
				if f.Type == FailureNodeOutage {
					if state, ok := nodeState[f.NodeID]; ok {
						state.Status = StatusDown
						state.ErrorRate = 100
						activeFailures[f.NodeID] = true
					}
				}
			}
		}

		// 1. Inject traffic
		for _, tp := range params.Traffic {
			if state, ok := nodeState[tp.NodeID]; ok {
				if state.Status != StatusDown {
					state.Requests += tp.RequestsPerSec
				}
			}
		}

		// 2. Propagate traffic and handle dependency failures
		propagateTrafficAndFailures(g, nodeState, activeFailures, params.Failures, t)

		// 3. Calculate metrics
		for _, node := range g.Nodes {
			state := nodeState[node.ID]

			if state.Status == StatusDown {
				state.CPUUsage = 0
				state.Latency = 0
				state.Requests = 0
				currentMetrics[node.ID] = *state
				continue
			}

			// Heuristic: Capacity = 100 RPS per replica
			capacity := float64(state.Replicas * 100)
			if capacity > 0 {
				state.CPUUsage = (state.Requests / capacity) * 100
			} else {
				state.CPUUsage = 0
			}

			if state.CPUUsage > 100 {
				state.CPUUsage = 100 + (state.CPUUsage-100)*0.1 // Saturated
			}

			// Latency heuristic: base 10ms + exponential increase with load
			state.Latency = 10 + math.Pow(state.CPUUsage/20, 2)

			// Memory heuristic: 20% base + 0.5% per RPS per replica
			state.MemoryUsage = 20 + (state.Requests/float64(state.Replicas))*0.5
			if state.MemoryUsage > 100 {
				state.MemoryUsage = 95 + 5*math.Tanh((state.MemoryUsage-95)/10)
			}

			// Simulation of HPA (Phase 2 Step 3)
			if state.CPUUsage > 70 {
				state.Replicas++
			} else if state.CPUUsage < 30 && state.Replicas > 1 {
				// Scale down slowly
				if t%5 == 0 {
					state.Replicas--
				}
			}

			currentMetrics[node.ID] = *state
		}

		result.Timeline[t] = TimestampResult{
			Timestamp: t,
			Nodes:     currentMetrics,
		}
	}

	return result
}

func propagateTrafficAndFailures(g *graph.Graph, nodeState map[string]*NodeMetrics, activeFailures map[string]bool, failures []FailureEvent, t int) {
	inDegree := make(map[string]int)
	dependsOn := make(map[string][]string)
	for _, node := range g.Nodes {
		inDegree[node.ID] = 0
	}
	for _, edge := range g.Edges {
		dependsOn[edge.Target] = append(dependsOn[edge.Target], edge.Source)
		inDegree[edge.Source]++
	}

	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		currState := nodeState[curr]

		for _, dep := range dependsOn[curr] {
			depState := nodeState[dep]

			// Check for network partition
			isPartitioned := false
			for _, f := range failures {
				if f.Type == FailureNetworkPartition && t >= f.StartTime && t <= f.EndTime {
					if (f.NodeID == curr && f.TargetID == dep) || (f.NodeID == dep && f.TargetID == curr) {
						isPartitioned = true
						break
					}
				}
			}

			if isPartitioned {
				currState.ErrorRate = math.Max(currState.ErrorRate, 50) // Partial failure
				currState.Status = StatusDegraded
			} else if depState.Status == StatusDown {
				// Dependency is down, so current node fails
				currState.ErrorRate = math.Max(currState.ErrorRate, 80) // High error rate
				currState.Status = StatusDegraded
			} else {
				// Propagate traffic
				depState.Requests += currState.Requests
			}

			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}
}
