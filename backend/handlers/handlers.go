package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/user/infratwin/backend/graph"
	"github.com/user/infratwin/backend/k8s"
	"github.com/user/infratwin/backend/parser"
)

type AnalyzeRequest struct {
	TerraformCode string `json:"terraformCode"`
}

func AnalyzeHandler(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a temporary directory to store the terraform code
	tmpDir, err := os.MkdirTemp("", "infratwin-*")
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

	c.JSON(http.StatusOK, gin.H{
		"graph":     infraGraph,
		"manifests": manifests,
		"config":    config,
	})
}
