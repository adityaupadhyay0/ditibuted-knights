package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Campaign struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	Name          string         `json:"name"`
	TerraformCode string         `json:"terraformCode"`
	Result        string         `gorm:"type:text" json:"result"` // JSON blob of simulation result
}

type StrongholdMetric struct {
	ID           uint      `gorm:"primaryKey"`
	CampaignID   uint      `gorm:"index"`
	Timestamp    time.Time `gorm:"index"`
	StrongholdID string    `gorm:"index"`
	CPUUsage     float64
	MemoryUsage  float64
	Latency      float64
	ErrorRate    float64
	Status       string
}

func InitDB() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Println("DB_URL not set, skipping DB init")
		return
	}

	var err error
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to DB, retrying in 5s... (%d/5)", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto migrate
	err = DB.AutoMigrate(&Campaign{}, &StrongholdMetric{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Create hypertable for StrongholdMetric if it's TimescaleDB
	DB.Exec("SELECT create_hypertable('stronghold_metrics', 'timestamp', if_not_exists => TRUE);")

	log.Println("Database initialized successfully")
}
