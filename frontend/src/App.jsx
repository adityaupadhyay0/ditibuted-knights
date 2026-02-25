import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import CytoscapeComponent from 'react-cytoscapejs';
import { Play, Code, Layout, Layers, AlertCircle, Zap, Activity, Clock, ShieldAlert, Trash2 } from 'lucide-react';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const DEFAULT_TF = `resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.public.id
}

module "app_server" {
  source = "./modules/app"
  vpc_id = aws_vpc.main.id
}
`;

function App() {
  const [terraformCode, setTerraformCode] = useState(DEFAULT_TF);
  const [graphData, setGraphData] = useState([]);
  const [manifests, setManifests] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Simulation states
  const [simulation, setSimulation] = useState(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [trafficSettings, setTrafficSettings] = useState({}); // nodeId -> rps
  const [failures, setFailures] = useState([]); // [{type, nodeId, startTime, endTime}]
  const [isSimulating, setIsSimulating] = useState(false);

  const cyRef = useRef(null);

  const analyzeCode = async (isSim = false) => {
    setLoading(true);
    setError(null);
    try {
      const simParams = isSim ? {
        durationSeconds: 30,
        traffic: Object.entries(trafficSettings).map(([nodeId, rps]) => ({ nodeId, rps: parseFloat(rps) })),
        failures: failures
      } : null;

      const response = await axios.post(`${API_URL}/analyze`, {
        terraformCode,
        simulationParams: simParams
      });

      const { graph, manifests, simulation: simResult } = response.data;

      const elements = [
        ...graph.nodes.map(n => ({ data: { id: n.id, label: n.label, type: n.type } })),
        ...graph.edges.map(e => ({ data: { id: e.id, source: e.source, target: e.target } }))
      ];

      setGraphData(elements);
      setManifests(manifests);
      setSimulation(simResult);
      if (simResult) {
        setIsSimulating(true);
        setCurrentTime(0);
      } else {
        setIsSimulating(false);
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    analyzeCode();
  }, []);

  useEffect(() => {
    if (cyRef.current && graphData.length > 0) {
      cyRef.current.elements().remove();
      cyRef.current.add(graphData);
      cyRef.current.layout({ name: 'cose', animate: true }).run();
      cyRef.current.fit();
    }
  }, [graphData]);

  // Update node styles based on simulation data
  useEffect(() => {
    if (cyRef.current && simulation && simulation.timeline[currentTime]) {
      const metrics = simulation.timeline[currentTime].nodes;
      Object.entries(metrics).forEach(([id, m]) => {
        const node = cyRef.current.$id(id);
        if (node) {
          // Color based on status
          let bgColor = '#4F46E5';
          if (m.status === 'down') bgColor = '#EF4444';
          else if (m.status === 'degraded') bgColor = '#F59E0B';
          else {
              const hue = Math.max(0, 240 - (m.cpuUsage * 2.4));
              bgColor = `hsl(${hue}, 70%, 50%)`;
          }

          node.style('background-color', bgColor);
          node.style('label', `${id}\n${m.status.toUpperCase()} | ${m.errorRate.toFixed(0)}% err`);

          if (m.status === 'down') {
              node.style('opacity', 0.6);
          } else {
              node.style('opacity', 1);
          }
        }
      });
    } else if (cyRef.current && !simulation) {
        cyRef.current.nodes().forEach(node => {
            node.style('background-color', node.data('type') === 'module' ? '#10B981' : '#4F46E5');
            node.style('label', node.data('label'));
            node.style('opacity', 1);
        });
    }
  }, [simulation, currentTime]);

  const style = [
    {
      selector: 'node',
      style: {
        'label': 'data(label)',
        'background-color': '#4F46E5',
        'color': '#fff',
        'text-valign': 'center',
        'text-halign': 'center',
        'font-size': '10px',
        'width': '140px',
        'height': '60px',
        'shape': 'round-rectangle',
        'text-wrap': 'wrap'
      }
    },
    {
      selector: 'node[type="module"]',
      style: {
        'background-color': '#10B981',
      }
    },
    {
      selector: 'edge',
      style: {
        'width': 2,
        'line-color': '#94A3B8',
        'target-arrow-color': '#94A3B8',
        'target-arrow-shape': 'triangle',
        'curve-style': 'bezier'
      }
    }
  ];

  const handleTrafficChange = (nodeId, val) => {
    setTrafficSettings(prev => ({ ...prev, [nodeId]: val }));
  };

  const addFailure = (nodeId) => {
    setFailures([...failures, {
        type: 'node_outage',
        nodeId: nodeId,
        startTime: 10,
        endTime: 20
    }]);
  };

  const removeFailure = (index) => {
    setFailures(failures.filter((_, i) => i !== index));
  };

  return (
    <div className="flex flex-col h-screen bg-gray-900 text-white font-sans">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center space-x-2">
          <Layers className="text-indigo-500" size={24} />
          <h1 className="text-xl font-bold">InfraTwin <span className="text-sm font-normal text-gray-400">v0.3.0</span></h1>
        </div>
        <div className="flex items-center space-x-4">
          <button
            onClick={() => analyzeCode(false)}
            disabled={loading}
            className="text-gray-300 hover:text-white px-3 py-1 text-sm transition"
          >
            Refresh Graph
          </button>
          <button
            onClick={() => analyzeCode(true)}
            disabled={loading}
            className="flex items-center space-x-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-800 px-4 py-2 rounded-md transition font-semibold shadow-lg shadow-indigo-500/20"
          >
            {loading ? <span>Analyzing...</span> : <><Zap size={16} /><span>Run Simulation</span></>}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 overflow-hidden">
        {/* Left Pane: Code, Traffic, & Chaos */}
        <div className="w-1/3 flex flex-col border-r border-gray-700 bg-gray-900 overflow-hidden">
          <div className="h-1/3 flex flex-col border-b border-gray-700">
            <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-700">
              <Code size={16} className="mr-2 text-gray-400" />
              <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">main.tf</span>
            </div>
            <textarea
              className="flex-1 p-4 bg-gray-900 text-gray-300 font-mono text-sm focus:outline-none resize-none"
              value={terraformCode}
              onChange={(e) => setTerraformCode(e.target.value)}
            />
          </div>

          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex border-b border-gray-700 bg-gray-800">
                <div className="flex-1 px-4 py-2 text-xs font-semibold text-gray-300 uppercase border-r border-gray-700 flex items-center">
                    <Activity size={14} className="mr-2" /> Traffic
                </div>
                <div className="flex-1 px-4 py-2 text-xs font-semibold text-gray-300 uppercase flex items-center">
                    <ShieldAlert size={14} className="mr-2" /> Chaos
                </div>
            </div>

            <div className="flex-1 flex overflow-hidden">
                {/* Traffic List */}
                <div className="flex-1 overflow-auto p-3 space-y-3 border-r border-gray-700">
                  {graphData.filter(el => el.data.id).map(node => (
                    <div key={node.data.id} className="bg-gray-800 p-2 rounded border border-gray-700">
                      <div className="flex justify-between items-center mb-1">
                        <span className="text-[10px] font-mono text-indigo-400 truncate w-24">{node.data.id}</span>
                        <span className="text-[10px] text-gray-400">{trafficSettings[node.data.id] || 0} RPS</span>
                      </div>
                      <input
                        type="range" min="0" max="500" step="10"
                        value={trafficSettings[node.data.id] || 0}
                        onChange={(e) => handleTrafficChange(node.data.id, e.target.value)}
                        className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-indigo-500"
                      />
                    </div>
                  ))}
                </div>

                {/* Chaos List */}
                <div className="flex-1 overflow-auto p-3 space-y-3">
                  <div className="mb-2">
                    <label className="text-[10px] text-gray-500 uppercase block mb-1">Inject Failure</label>
                    <select
                        onChange={(e) => e.target.value && addFailure(e.target.value)}
                        className="w-full bg-gray-800 border border-gray-700 text-[10px] rounded p-1 outline-none"
                        value=""
                    >
                        <option value="">Select Node...</option>
                        {graphData.filter(el => el.data.id).map(node => (
                            <option key={node.data.id} value={node.data.id}>{node.data.id}</option>
                        ))}
                    </select>
                  </div>

                  {failures.map((f, i) => (
                    <div key={i} className="bg-red-900/20 p-2 rounded border border-red-500/30">
                        <div className="flex justify-between items-start mb-1">
                            <span className="text-[10px] font-bold text-red-400 uppercase">{f.type}</span>
                            <button onClick={() => removeFailure(i)}><Trash2 size={12} className="text-gray-500 hover:text-red-400"/></button>
                        </div>
                        <div className="text-[10px] text-gray-300 truncate mb-1">{f.nodeId}</div>
                        <div className="flex items-center space-x-2">
                            <input
                                type="number" value={f.startTime}
                                onChange={(e) => {
                                    const newF = [...failures];
                                    newF[i].startTime = parseInt(e.target.value);
                                    setFailures(newF);
                                }}
                                className="w-10 bg-black/40 text-[9px] p-0.5 rounded outline-none"
                            />
                            <span className="text-[9px] text-gray-500">to</span>
                            <input
                                type="number" value={f.endTime}
                                onChange={(e) => {
                                    const newF = [...failures];
                                    newF[i].endTime = parseInt(e.target.value);
                                    setFailures(newF);
                                }}
                                className="w-10 bg-black/40 text-[9px] p-0.5 rounded outline-none"
                            />
                        </div>
                    </div>
                  ))}
                </div>
            </div>
          </div>
        </div>

        {/* Middle Pane: Graph & Timeline */}
        <div className="flex-1 flex flex-col relative bg-gray-950">
          <div className="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700">
            <div className="flex items-center">
              <Layout size={16} className="mr-2 text-gray-400" />
              <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">Topology Simulation</span>
            </div>
            {isSimulating && (
                <div className="flex items-center space-x-2 text-xs font-mono text-indigo-400">
                    <Clock size={14} />
                    <span>T + {currentTime}s</span>
                </div>
            )}
          </div>
          <div className="flex-1">
            <CytoscapeComponent
              elements={[]}
              style={{ width: '100%', height: '100%' }}
              stylesheet={style}
              cy={(cy) => { cyRef.current = cy }}
            />
          </div>

          {/* Timeline Slider */}
          {isSimulating && simulation && (
            <div className="absolute bottom-6 left-1/2 -translate-x-1/2 w-3/4 bg-gray-800/90 backdrop-blur p-4 rounded-xl border border-indigo-500/30 shadow-2xl">
                <div className="flex items-center space-x-4">
                    <button onClick={() => setIsSimulating(!isSimulating)} className="text-indigo-400 hover:text-indigo-300">
                        <Activity size={24} />
                    </button>
                    <input
                        type="range" min="0" max={simulation.timeline.length - 1}
                        value={currentTime}
                        onChange={(e) => setCurrentTime(parseInt(e.target.value))}
                        className="flex-1 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-indigo-500"
                    />
                </div>
                <div className="flex justify-between mt-2 text-[10px] text-gray-500 font-mono uppercase tracking-tighter">
                    <span>Start</span>
                    <span>Simulation Progress ({currentTime}s / {simulation.timeline.length - 1}s)</span>
                    <span>End</span>
                </div>
            </div>
          )}

          {error && (
            <div className="absolute top-4 left-4 right-4 bg-red-900/80 backdrop-blur border border-red-500 text-red-200 px-4 py-3 rounded-md flex items-center space-x-3 z-50 shadow-2xl">
              <AlertCircle size={20} />
              <span>{error}</span>
            </div>
          )}
        </div>

        {/* Right Pane: Stats & Manifests */}
        <div className="w-1/4 flex flex-col border-l border-gray-700 bg-gray-800">
           <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-700">
            <Activity size={16} className="mr-2 text-gray-400" />
            <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">Node Stats</span>
          </div>
          <div className="p-4 flex-1 overflow-auto">
            {simulation && simulation.timeline[currentTime] ? (
                <div className="space-y-4">
                    {Object.entries(simulation.timeline[currentTime].nodes).map(([id, m]) => (
                        <div key={id} className={`p-3 rounded border ${m.status === 'down' ? 'bg-red-900/20 border-red-500/50' : 'bg-gray-900 border-gray-700'}`}>
                            <div className="flex justify-between items-center mb-2">
                                <h4 className={`text-xs font-bold truncate ${m.status === 'down' ? 'text-red-400' : 'text-indigo-400'}`}>{id}</h4>
                                <span className={`text-[8px] px-1 rounded uppercase font-bold ${
                                    m.status === 'down' ? 'bg-red-500 text-white' :
                                    m.status === 'degraded' ? 'bg-amber-500 text-white' : 'bg-green-500 text-white'
                                }`}>{m.status}</span>
                            </div>
                            <div className="grid grid-cols-2 gap-2 text-[10px]">
                                <div>
                                    <div className="text-gray-500 uppercase">CPU Usage</div>
                                    <div className={`font-mono ${m.cpuUsage > 80 ? 'text-red-400' : 'text-green-400'}`}>{m.cpuUsage.toFixed(1)}%</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase">Latency</div>
                                    <div className="font-mono text-gray-300">{m.latency.toFixed(1)}ms</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase">Replicas</div>
                                    <div className="font-mono text-gray-300">{m.replicas}</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase">Error Rate</div>
                                    <div className={`font-mono ${m.errorRate > 0 ? 'text-red-400 font-bold' : 'text-gray-300'}`}>{m.errorRate.toFixed(1)}%</div>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center text-gray-500 mt-20">
                    <Activity size={48} className="mx-auto mb-4 opacity-20" />
                    <p className="text-sm">Run simulation to see real-time metrics.</p>
                </div>
            )}
          </div>

          <div className="h-1/4 border-t border-gray-700 flex flex-col">
            <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-700">
                <Layers size={16} className="mr-2 text-gray-400" />
                <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">K8s Manifests</span>
            </div>
            <pre className="flex-1 p-4 overflow-auto text-[10px] text-gray-400 font-mono">
                {manifests || "# No manifests generated yet."}
            </pre>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;
