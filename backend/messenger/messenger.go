package messenger

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/user/distributed-knights/backend/graph"
	"github.com/user/distributed-knights/backend/parser"
	"github.com/user/distributed-knights/backend/simulation"
)

var NC *nats.Conn

type SimulationTask struct {
	Config           *parser.InfraConfig        `json:"config"`
	Graph            *graph.Graph               `json:"graph"`
	SimulationParams *simulation.CampaignParams `json:"simulationParams"`
}

func InitNATS() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	var err error
	for i := 0; i < 5; i++ {
		NC, err = nats.Connect(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to NATS, retrying in 5s... (%d/5)", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Printf("Could not connect to NATS: %v. Running in local mode.", err)
	} else {
		log.Println("Connected to NATS (The Messenger's Guild)")
	}
}

func StartWorker() {
	if NC == nil {
		return
	}

	_, err := NC.Subscribe("simulation.tasks", func(m *nats.Msg) {
		var task SimulationTask
		if err := json.Unmarshal(m.Data, &task); err != nil {
			log.Printf("Error unmarshaling task: %v", err)
			return
		}

		log.Printf("Worker received campaign task for %d strongholds", len(task.Graph.Nodes))
		result := simulation.RunCampaign(task.Config, task.Graph, task.SimulationParams)

		resultData, _ := json.Marshal(result)
		m.Respond(resultData)
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to simulation.tasks: %v", err)
	}

	log.Println("Simulation Worker started and listening for tasks")
}

func RequestSimulation(task SimulationTask) (*simulation.CampaignResult, error) {
	if NC == nil {
		// Fallback to local execution if NATS is not available
		log.Println("NATS not available, falling back to local simulation")
		return simulation.RunCampaign(task.Config, task.Graph, task.SimulationParams), nil
	}

	taskData, _ := json.Marshal(task)

	// We use Request for synchronous simulation in this case,
	// but the architecture allows for async workers.
	msg, err := NC.Request("simulation.tasks", taskData, 10*time.Second)
	if err != nil {
		return nil, err
	}

	var result simulation.CampaignResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
