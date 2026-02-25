package simulation

import (
	"math"

	"github.com/user/distributed-knights/backend/graph"
	"github.com/user/distributed-knights/backend/parser"
)

type SiegeType string

const (
	SiegeDragonStrike     SiegeType = "dragon_strike"      // Node outage
	SiegeNetworkPartition SiegeType = "network_partition"  // Network split
	SiegePlague           SiegeType = "plague"            // Cascading errors
	SiegeFamine           SiegeType = "famine"            // Resource exhaustion
)

type WeatherCondition string

const (
	WeatherClear  WeatherCondition = "clear"
	WeatherStorm  WeatherCondition = "storm" // Increases latency
	WeatherFog    WeatherCondition = "fog"   // Increases error rate
	WeatherBlizzard WeatherCondition = "blizzard" // Severe latency and errors
)

type SiegeEvent struct {
	Type      SiegeType `json:"type"`
	NodeID    string    `json:"nodeId"`
	TargetID  string    `json:"targetId,omitempty"` // For network partitions
	StartTime int       `json:"startTime"`
	EndTime   int       `json:"endTime"`
}

type MessengerPattern struct {
	NodeID         string  `json:"nodeId"`
	RequestsPerSec float64 `json:"rps"`
}

type CampaignParams struct {
	DurationSeconds int                `json:"durationSeconds"`
	Traffic         []MessengerPattern `json:"traffic"`
	Failures        []SiegeEvent       `json:"failures"`
	Weather         WeatherCondition   `json:"weather"`
}

type CampaignResult struct {
	Timeline []TimestampResult `json:"timeline"`
	Scrolls  []string          `json:"scrolls"` // Scribe flavor text
}

type TimestampResult struct {
	Timestamp int                    `json:"timestamp"`
	Nodes     map[string]NodeMetrics `json:"nodes"`
	Events    []string               `json:"events"` // Significant events at this step
}

type StrongholdStatus string

const (
	StatusStanding StrongholdStatus = "up"
	StatusFallen   StrongholdStatus = "down"
	StatusBesieged StrongholdStatus = "degraded"
)

type NodeMetrics struct {
	CPUUsage    float64          `json:"cpuUsage"`    // Percentage 0-100
	MemoryUsage float64          `json:"memoryUsage"` // Percentage 0-100
	Latency     float64          `json:"latency"`     // Milliseconds
	Replicas    int              `json:"replicas"`
	Requests    float64          `json:"requests"`    // Requests handled at this node
	ErrorRate   float64          `json:"errorRate"`   // Percentage 0-100
	Status      StrongholdStatus `json:"status"`
}

func RunCampaign(config *parser.InfraConfig, g *graph.Graph, params *CampaignParams) *CampaignResult {
	if params.DurationSeconds <= 0 {
		params.DurationSeconds = 10
	}

	result := &CampaignResult{
		Timeline: make([]TimestampResult, params.DurationSeconds),
		Scrolls:  []string{"Hear ye, hear ye! A new campaign begins in the realm of Distributed Knights!"},
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
			nodeState[id].Status = StatusStanding
			nodeState[id].ErrorRate = 0
		}

		// 0.1 Apply active failures (Siege Events)
		activeFailures := make(map[string]bool)
		stepEvents := []string{}
		for _, f := range params.Failures {
			if t == f.StartTime {
				msg := scribeEvent(f)
				stepEvents = append(stepEvents, msg)
				result.Scrolls = append(result.Scrolls, msg)
			}

			if t >= f.StartTime && t <= f.EndTime {
				if state, ok := nodeState[f.NodeID]; ok {
					switch f.Type {
					case SiegeDragonStrike:
						state.Status = StatusFallen
						state.ErrorRate = 100
						activeFailures[f.NodeID] = true
					case SiegePlague:
						state.Status = StatusBesieged
						state.ErrorRate = math.Max(state.ErrorRate, 40)
					case SiegeFamine:
						state.Status = StatusBesieged
						state.CPUUsage = math.Max(state.CPUUsage, 90)
					}
				}
			}
		}

		// 1. Inject traffic
		for _, tp := range params.Traffic {
			if state, ok := nodeState[tp.NodeID]; ok {
				if state.Status != StatusFallen {
					state.Requests += tp.RequestsPerSec
				}
			}
		}

		// 2. Propagate traffic and handle dependency failures
		propagateTrafficAndFailures(g, nodeState, activeFailures, params.Failures, t)

		// 3. Calculate metrics
		for _, node := range g.Nodes {
			state := nodeState[node.ID]

			if state.Status == StatusFallen {
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

			// Weather impact
			weatherLatencyMod := 1.0
			weatherErrorMod := 0.0
			switch params.Weather {
			case WeatherStorm:
				weatherLatencyMod = 2.5
			case WeatherFog:
				weatherErrorMod = 5.0
			case WeatherBlizzard:
				weatherLatencyMod = 5.0
				weatherErrorMod = 15.0
			}

			if state.CPUUsage > 100 {
				state.CPUUsage = 100 + (state.CPUUsage-100)*0.1 // Saturated
			}

			// Latency heuristic: base 10ms + exponential increase with load
			state.Latency = (10 + math.Pow(state.CPUUsage/20, 2)) * weatherLatencyMod
			state.ErrorRate = math.Max(state.ErrorRate, weatherErrorMod)

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
			Events:    stepEvents,
		}
	}

	return result
}

func scribeEvent(f SiegeEvent) string {
	switch f.Type {
	case SiegeDragonStrike:
		return "A dragon has descended upon " + f.NodeID + "! The stronghold has fallen!"
	case SiegeNetworkPartition:
		return "A great rift has opened between " + f.NodeID + " and " + f.TargetID + "! Messengers cannot pass!"
	case SiegePlague:
		return "A mysterious plague is spreading through the garrison of " + f.NodeID + ". Chaos ensues!"
	case SiegeFamine:
		return "Rations are running low at " + f.NodeID + ". The battalions are exhausted!"
	default:
		return "An unknown calamity has struck " + f.NodeID + "!"
	}
}

func propagateTrafficAndFailures(g *graph.Graph, nodeState map[string]*NodeMetrics, activeFailures map[string]bool, failures []SiegeEvent, t int) {
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
				if f.Type == SiegeNetworkPartition && t >= f.StartTime && t <= f.EndTime {
					if (f.NodeID == curr && f.TargetID == dep) || (f.NodeID == dep && f.TargetID == curr) {
						isPartitioned = true
						break
					}
				}
			}

			if isPartitioned {
				currState.ErrorRate = math.Max(currState.ErrorRate, 50) // Partial failure
				currState.Status = StatusBesieged
			} else if depState.Status == StatusFallen {
				// Dependency is down, so current node fails
				currState.ErrorRate = math.Max(currState.ErrorRate, 80) // High error rate
				currState.Status = StatusBesieged
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
