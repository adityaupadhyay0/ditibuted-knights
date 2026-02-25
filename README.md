# InfraTwin (distributed-knights)

A distributed systems digital twin platform for pre-production validation and controlled deployment.

## Current Status: Phase 2 (Load Simulation)
The current implementation provides a deterministic simulation engine for infrastructure defined via Terraform.

### Features
- **Terraform Parser**: Extracts resources and dependencies from HCL files.
- **Dependency DAG**: Builds a graph of infrastructure components.
- **Load Simulation Engine**: Timestep-based simulation of traffic, latency, and resource usage.
- **Autoscaling Modeling**: Simulates Kubernetes HPA behavior under varying load.
- **Traffic Injection UI**: Interactive sliders to inject traffic at any node and see the ripple effect.
- **Topology Visualization**: Interactive 2D graph of your infrastructure using Cytoscape.js.
- **K8s Manifest Generation**: Automatically generates skeleton Kubernetes manifests based on Terraform resources.
- **Dockerized Environment**: Easy to spin up using Docker Compose.

## Getting Started

### Prerequisites
- Docker
- Docker Compose

### Running the Platform
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
- `samples/`: Sample Terraform files for testing.
