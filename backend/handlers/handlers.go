package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/distributed-knights/backend/analysis"
	"github.com/user/distributed-knights/backend/db"
	"github.com/user/distributed-knights/backend/graph"
	"github.com/user/distributed-knights/backend/k8s"
	"github.com/user/distributed-knights/backend/messenger"
	"github.com/user/distributed-knights/backend/mirror"
	"github.com/user/distributed-knights/backend/parser"
	"github.com/user/distributed-knights/backend/simulation"
)

type AnalyzeRequest struct {
	TerraformCode    string                     `json:"terraformCode"`
	SimulationParams *simulation.CampaignParams `json:"simulationParams"`
}

func AnalyzeHandler(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a temporary directory to store the terraform code
	tmpDir, err := os.MkdirTemp("", "distributed-knights-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp dir"})
		return
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "main.tf")
	if err := os.WriteFile(tmpFile, []byte(req.TerraformCode), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write temp file"})
		return
	}

	config, err := parser.ParseTerraform(tmpDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	infraGraph := graph.BuildGraph(config)
	manifests := k8s.GenerateManifests(config)

	// Architectural Analysis (The Royal Treasury & Castle Guard)
	treasuryReport := analysis.AuditTreasury(config)
	guardReport := analysis.AuditCastle(config)
	mirrorReport := mirror.InspectMirror(config)

	var simResult *simulation.CampaignResult
	if req.SimulationParams != nil {
		var err error
		simResult, err = messenger.RequestSimulation(messenger.SimulationTask{
			Config:           config,
			Graph:            infraGraph,
			SimulationParams: req.SimulationParams,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "simulation failed: " + err.Error()})
			return
		}

		// Save Campaign to The Chronicles
		if db.DB != nil {
			resultJSON, _ := json.Marshal(simResult)
			campaign := db.Campaign{
				Name:          "Campaign " + time.Now().Format("2006-01-02 15:04:05"),
				TerraformCode: req.TerraformCode,
				Result:        string(resultJSON),
			}
			db.DB.Create(&campaign)

			// Extract and save metrics for TimescaleDB
			for _, ts := range simResult.Timeline {
				baseTime := time.Now().Truncate(time.Hour) // Just a base for simulation time
				for nodeID, metrics := range ts.Nodes {
					db.DB.Create(&db.StrongholdMetric{
						CampaignID:   campaign.ID,
						Timestamp:    baseTime.Add(time.Duration(ts.Timestamp) * time.Second),
						StrongholdID: nodeID,
						CPUUsage:     metrics.CPUUsage,
						MemoryUsage:  metrics.MemoryUsage,
						Latency:      metrics.Latency,
						ErrorRate:    metrics.ErrorRate,
						Status:       string(metrics.Status),
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"graph":      infraGraph,
		"manifests":  manifests,
		"config":     config,
		"simulation": simResult,
		"treasury":   treasuryReport,
		"guard":      guardReport,
		"mirror":     mirrorReport,
	})
}

func ListCampaignsHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	var campaigns []db.Campaign
	db.DB.Order("created_at desc").Limit(10).Find(&campaigns)
	c.JSON(http.StatusOK, campaigns)
}

func GetCampaignHandler(c *gin.Context) {
	if db.DB == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "database not connected"})
		return
	}

	id := c.Param("id")
	var campaign db.Campaign
	if err := db.DB.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}
	c.JSON(http.StatusOK, campaign)
}
