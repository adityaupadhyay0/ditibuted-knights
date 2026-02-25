package mirror

import (
	"github.com/user/distributed-knights/backend/parser"
)

type DriftType string

const (
	DriftMissing   DriftType = "missing"   // Resource in Decree but not in Reality
	DriftExtra     DriftType = "extra"     // Resource in Reality but not in Decree
	DriftModified  DriftType = "modified"  // Properties differ
)

type DriftFinding struct {
	StrongholdID string    `json:"strongholdId"`
	Type         DriftType `json:"type"`
	Description  string    `json:"description"`
}

type MirrorReport struct {
	Drifts []DriftFinding `json:"drifts"`
	InSync bool           `json:"inSync"`
}

func InspectMirror(config *parser.InfraConfig) *MirrorReport {
	report := &MirrorReport{
		Drifts: []DriftFinding{},
		InSync: true,
	}

	// For simulation/demo purposes, we'll introduce some random "drift"
	// based on specific resource names to make it deterministic for the user.
	for _, res := range config.Resources {
		id := res.Type + "." + res.Name

		if res.Name == "web" {
			report.Drifts = append(report.Drifts, DriftFinding{
				StrongholdID: id,
				Type:         DriftModified,
				Description:  "Garrison size (instance_type) in reality is 't3.large', but Decree says 't2.micro'.",
			})
			report.InSync = false
		}
	}

	return report
}
