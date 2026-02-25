# Distributed Knights

A distributed systems digital twin platform for pre-production validation and controlled deployment.

## Current Status: Phase 3 (Siege & Failure Modeling)
The current implementation provides a deterministic simulation engine for infrastructure defined via "Architectural Decrees" (Terraform), including failure propagation.

### Features
- **Decree Parser**: Extracts resources and dependencies from HCL files.
- **Dependency DAG**: Builds a graph of infrastructure strongholds.
- **Messenger Simulation Engine**: Timestep-based simulation of messenger traffic, latency, and resource usage.
- **Siege Engineering**: Inject stronghold outages and network partitions to see failure propagation.
- **Autoscaling Modeling**: Simulates Kubernetes HPA behavior under varying load.
- **Messenger Injection UI**: Interactive sliders to inject messengers at any stronghold.
- **War Map Visualization**: Interactive 2D graph with status indicators (Up/Down/Degraded).
- **K8s Manifest Generation**: Automatically generates skeleton Kubernetes manifests based on Terraform resources.
- **Dockerized Environment**: Easy to spin up using Docker Compose.

## Getting Started

### Prerequisites
- Docker
- Docker Compose

### Launching the Campaign
1. Clone the repository.
2. Run the following command:
   ```bash
   docker-compose up --build
   ```
3. Open your browser and navigate to `http://localhost:3000`.

## Technical Stack
- **Backend**: Go with Gin and HCL v2.
- **Frontend**: React with Cytoscape.js and Tailwind CSS.
- **Infrastructure**: Docker for orchestration.

## Project Structure
- `backend/`: Go source code for the API and simulation engine.
- `frontend/`: React source code for the dashboard.
- `samples/`: Sample "Decrees" (Terraform files) for testing.
