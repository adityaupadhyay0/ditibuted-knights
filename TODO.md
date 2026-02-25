# Future Upgrades & Roadmap

## Phase 2 – Deterministic Load Simulation
- [x] **Simplified Latency Model**: Implement a model to predict latency between nodes in the graph.
- [x] **CPU/Memory Saturation Modeling**: Simulate resource pressure based on traffic patterns.
- [x] **Basic Autoscaling Simulation**: Predict when K8s Horizontal Pod Autoscalers (HPA) would trigger.
- [x] **Traffic Injection**: Allow users to define traffic patterns in the UI to see impact on the graph.

## Phase 3 – Failure & Edge Modeling
- [ ] **Zone Outage Simulation**: Simulate the impact of an entire availability zone going down.
- [ ] **Network Partition Simulation**: Model split-brain scenarios.
- [ ] **Edge Node Latency Validation**: Integrate real-world latency measurements from edge agents.
- [ ] **Chaos Engineering Integration**: Trigger failure events directly from the dashboard.

## Technical Improvements
- [ ] **Persistent Storage**: Integrate PostgreSQL for metadata and TimescaleDB for simulation metrics.
- [ ] **NATS/Kafka Integration**: Scale the simulation plane using a distributed message bus.
- [ ] **Advanced HCL Parsing**: Support more complex Terraform expressions and remote modules.
- [ ] **Real-world Validation**: Compare simulation results with actual production metrics (drift detection).
