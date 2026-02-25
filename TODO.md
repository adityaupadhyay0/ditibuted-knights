# Future Campaigns & Quest Log

## Phase 2 – Deterministic Messenger Simulation
- [x] **Simplified Latency Model**: Predict delay between strongholds in the realm.
- [x] **Garrison Saturation Modeling**: Simulate resource pressure based on messenger patterns.
- [x] **Basic Battalions Simulation**: Predict when K8s HPA would summon more reinforcements.
- [x] **Messenger Injection**: Allow users to define messenger patterns to see impact on the realm.

## Phase 3 – Siege & Edge Modeling
- [x] **Realm Outage Simulation**: Simulate the impact of an entire availability zone falling.
- [x] **Network Partition Simulation**: Model split-realm scenarios.
- [ ] **Edge Stronghold Latency Validation**: Integrate real-world latency measurements from edge scouts.
- [x] **Siege Engineering Integration**: Trigger assault events directly from the War Map.

## Technical Improvements
- [x] **The Chronicles**: Integrate PostgreSQL for metadata and TimescaleDB for campaign metrics.
- [x] **The Messenger's Guild**: Scale the campaign plane using a distributed message bus (NATS/Kafka).
- [ ] **Advanced Architectural Parsing**: Support more complex Decree expressions and remote modules.
- [x] **The Mirror Realm**: Compare campaign results with actual production metrics (drift detection).
