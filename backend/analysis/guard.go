package analysis

import (
	"strings"
	"github.com/user/distributed-knights/backend/parser"
)

type GuardVulnerability struct {
	StrongholdID string `json:"strongholdId"`
	Severity     string `json:"severity"` // Low, Medium, High, Critical
	Finding      string `json:"finding"`
	Advice       string `json:"advice"`
}

type GuardReport struct {
	Vulnerabilities []GuardVulnerability `json:"vulnerabilities"`
	SecurityScore   int                  `json:"securityScore"` // 0-100
}

func AuditCastle(config *parser.InfraConfig) *GuardReport {
	report := &GuardReport{
		Vulnerabilities: []GuardVulnerability{},
		SecurityScore:   100,
	}

	for _, res := range config.Resources {
		id := res.Type + "." + res.Name

		// Mock security checks
		if res.Type == "aws_security_group" || res.Type == "aws_instance" {
			// In a real implementation, we'd parse the 'res.Properties'
			// For this 10x demo, we'll simulate findings based on resource names or types
			if strings.Contains(res.Name, "public") || strings.Contains(res.Name, "open") {
				report.Vulnerabilities = append(report.Vulnerabilities, GuardVulnerability{
					StrongholdID: id,
					Severity:     "High",
					Finding:      "The gate is wide open to the public realm (Port 22/3389).",
					Advice:       "Restrict access to known allies only.",
				})
				report.SecurityScore -= 20
			}
		}

		if strings.Contains(res.Type, "db") && !strings.Contains(res.Name, "private") {
			report.Vulnerabilities = append(report.Vulnerabilities, GuardVulnerability{
				StrongholdID: id,
				Severity:     "Critical",
				Finding:      "The Royal Treasury (Database) is not within the inner keep.",
				Advice:       "Move the database to a private subnet.",
			})
			report.SecurityScore -= 30
		}
	}

	if report.SecurityScore < 0 {
		report.SecurityScore = 0
	}

	return report
}
