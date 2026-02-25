package analysis

import (
	"github.com/user/distributed-knights/backend/parser"
)

type TreasuryReport struct {
	TotalMonthlyGold float64            `json:"totalMonthlyGold"`
	ResourceCosts    map[string]float64 `json:"resourceCosts"`
}

var goldPrices = map[string]float64{
	"aws_instance": 20.0,
	"aws_db_instance": 50.0,
	"aws_vpc": 0.0,
	"aws_subnet": 0.0,
	"aws_elb": 15.0,
	"aws_alb": 18.0,
	"google_compute_instance": 18.0,
	"google_container_cluster": 100.0,
}

func AuditTreasury(config *parser.InfraConfig) *TreasuryReport {
	report := &TreasuryReport{
		ResourceCosts: make(map[string]float64),
	}

	for _, res := range config.Resources {
		price := goldPrices[res.Type]
		if price == 0 && res.Type != "aws_vpc" && res.Type != "aws_subnet" {
			price = 5.0 // Default for unknown resources
		}

		id := res.Type + "." + res.Name
		report.ResourceCosts[id] = price
		report.TotalMonthlyGold += price
	}

	for _, mod := range config.Modules {
		price := 30.0 // Default for modules
		id := "module." + mod.Name
		report.ResourceCosts[id] = price
		report.TotalMonthlyGold += price
	}

	return report
}
