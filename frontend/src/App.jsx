import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import CytoscapeComponent from 'react-cytoscapejs';
import { Play, Code, Layout, Layers, AlertCircle, Zap, Activity, Clock, ShieldAlert, Trash2, Shield, Sword, Castle, Flame, Scroll, Skull } from 'lucide-react';

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

  // Architectural Analysis States
  const [treasury, setTreasury] = useState(null);
  const [guard, setGuard] = useState(null);
  const [mirror, setMirror] = useState(null);
  const [campaigns, setCampaigns] = useState([]);
  const [activeTab, setActiveTab] = useState('decree'); // decree, chronicles, treasury, guard

  // Simulation states
  const [simulation, setSimulation] = useState(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [trafficSettings, setTrafficSettings] = useState({}); // nodeId -> rps
  const [failures, setFailures] = useState([]); // [{type, nodeId, startTime, endTime}]
  const [weather, setWeather] = useState('clear');
  const [isSimulating, setIsSimulating] = useState(false);

  const cyRef = useRef(null);

  const analyzeCode = async (isSim = false) => {
    setLoading(true);
    setError(null);
    try {
      const simParams = isSim ? {
        durationSeconds: 30,
        traffic: Object.entries(trafficSettings).map(([nodeId, rps]) => ({ nodeId, rps: parseFloat(rps) })),
        failures: failures,
        weather: weather
      } : null;

      const response = await axios.post(`${API_URL}/analyze`, {
        terraformCode,
        simulationParams: simParams
      });

      const { graph, manifests, simulation: simResult, treasury: treasuryResult, guard: guardResult, mirror: mirrorResult } = response.data;

      const elements = [
        ...graph.nodes.map(n => ({ data: { id: n.id, label: n.label, type: n.type } })),
        ...graph.edges.map(e => ({ data: { id: e.id, source: e.source, target: e.target } }))
      ];

      setGraphData(elements);
      setManifests(manifests);
      setSimulation(simResult);
      setTreasury(treasuryResult);
      setGuard(guardResult);
      setMirror(mirrorResult);
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

  const fetchCampaigns = async () => {
    try {
      const response = await axios.get(`${API_URL}/campaigns`);
      setCampaigns(response.data);
    } catch (err) {
      console.error("Failed to fetch chronicles", err);
    }
  };

  useEffect(() => {
    analyzeCode();
    fetchCampaigns();
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
        'background-color': '#1e293b',
        'color': '#fde68a',
        'text-valign': 'bottom',
        'text-halign': 'center',
        'font-size': '8px',
        'width': '50px',
        'height': '50px',
        'shape': 'hexagon',
        'border-width': 2,
        'border-color': '#b45309',
        'text-margin-y': 5,
      }
    },
    {
      selector: 'node[type="module"]',
      style: {
        'background-color': '#047857',
        'shape': 'diamond',
        'width': '60px',
        'height': '60px',
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

  const addFailure = (nodeId, type = 'dragon_strike') => {
    setFailures([...failures, {
        type: type,
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
          <Shield className="text-amber-500" size={24} />
          <h1 className="text-xl font-bold tracking-tight">Distributed Knights <span className="text-sm font-normal text-gray-400">v1.0.0-omega</span></h1>
        </div>
        <div className="flex items-center space-x-4">
          <button
            onClick={() => analyzeCode(false)}
            disabled={loading}
            className="text-gray-300 hover:text-white px-3 py-1 text-sm transition font-serif italic"
          >
            Survey Lands
          </button>
          <button
            onClick={() => analyzeCode(true)}
            disabled={loading}
            className="flex items-center space-x-2 bg-amber-700 hover:bg-amber-600 disabled:bg-gray-700 px-4 py-2 rounded-md transition font-bold shadow-lg shadow-amber-900/40 border border-amber-500/30"
          >
            {loading ? <span>Enacting Decrees...</span> : <><Sword size={16} /><span>Launch Campaign</span></>}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 overflow-hidden">
        {/* Left Pane: The Great Hall */}
        <div className="w-1/3 flex flex-col border-r border-gray-800 bg-gray-900 overflow-hidden">
          {/* Navigation Tabs */}
          <div className="flex bg-gray-900 border-b border-gray-800">
             {[
               { id: 'decree', icon: Scroll, label: 'Decrees' },
               { id: 'chronicles', icon: Clock, label: 'Chronicles' },
               { id: 'treasury', icon: Zap, label: 'Treasury' },
               { id: 'guard', icon: Shield, label: 'Guard' }
             ].map(tab => (
               <button
                 key={tab.id}
                 onClick={() => setActiveTab(tab.id)}
                 className={`flex-1 flex flex-col items-center py-2 transition ${activeTab === tab.id ? 'bg-gray-800 text-amber-500 border-b-2 border-amber-500' : 'text-gray-500 hover:text-gray-300'}`}
               >
                 <tab.icon size={16} />
                 <span className="text-[9px] uppercase font-bold mt-1">{tab.label}</span>
               </button>
             ))}
          </div>

          <div className="h-1/3 flex flex-col border-b border-gray-800">
            {activeTab === 'decree' && (
              <textarea
                className="flex-1 p-4 bg-gray-950 text-amber-100/80 font-mono text-sm focus:outline-none resize-none"
                value={terraformCode}
                onChange={(e) => setTerraformCode(e.target.value)}
                placeholder="Write your architectural decrees here..."
              />
            )}
            {activeTab === 'chronicles' && (
              <div className="flex-1 overflow-auto bg-gray-950 p-2 space-y-2">
                {campaigns.map(c => (
                  <div key={c.id} className="p-2 border border-gray-800 hover:border-amber-900 rounded bg-gray-900/50 cursor-pointer group">
                    <div className="text-[10px] text-amber-500 font-bold">{c.name}</div>
                    <div className="text-[8px] text-gray-500">{new Date(c.createdAt).toLocaleString()}</div>
                  </div>
                ))}
              </div>
            )}
            {activeTab === 'treasury' && treasury && (
              <div className="flex-1 overflow-auto bg-gray-950 p-4">
                <div className="text-center mb-4">
                  <div className="text-[10px] text-gray-500 uppercase tracking-widest">Total Monthly Tithe</div>
                  <div className="text-3xl font-serif text-amber-500">{treasury.totalMonthlyGold.toFixed(0)} <span className="text-sm">Gold</span></div>
                </div>
                <div className="space-y-2">
                   {Object.entries(treasury.resourceCosts).map(([id, cost]) => (
                     <div key={id} className="flex justify-between text-[10px] border-b border-gray-800 pb-1">
                       <span className="text-gray-400 truncate w-32">{id}</span>
                       <span className="text-amber-200">{cost.toFixed(1)} Gold</span>
                     </div>
                   ))}
                </div>
              </div>
            )}
            {activeTab === 'guard' && guard && (
              <div className="flex-1 overflow-auto bg-gray-950 p-4">
                <div className="flex items-center justify-between mb-4 bg-gray-900 p-2 rounded border border-amber-900/30">
                  <div className="text-[10px] text-gray-400 uppercase">Defensive Rating</div>
                  <div className={`text-xl font-bold ${guard.securityScore > 80 ? 'text-green-500' : 'text-orange-500'}`}>{guard.securityScore}%</div>
                </div>
                <div className="space-y-3">
                  {guard.vulnerabilities.map((v, i) => (
                    <div key={i} className="p-2 border-l-2 border-orange-500 bg-orange-950/10 rounded-r">
                      <div className="flex justify-between items-center mb-1">
                         <span className="text-[9px] font-bold text-orange-400 uppercase">{v.severity} finding</span>
                      </div>
                      <div className="text-[10px] text-gray-200 mb-1">{v.finding}</div>
                      <div className="text-[9px] text-gray-500 italic">Advice: {v.advice}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex border-b border-gray-800 bg-gray-900">
                <div className="flex-1 px-4 py-2 text-xs font-semibold text-gray-300 uppercase border-r border-gray-800 flex items-center">
                    <Activity size={14} className="mr-2 text-blue-400" /> Messengers
                </div>
                <div className="flex-1 px-4 py-2 text-xs font-semibold text-gray-300 uppercase flex items-center">
                    <Flame size={14} className="mr-2 text-orange-500" /> Siege
                </div>
            </div>

            <div className="flex-1 flex overflow-hidden">
                {/* Messengers List */}
                <div className="flex-1 overflow-auto p-3 space-y-3 border-r border-gray-700">
                  {graphData.filter(el => el.data.id).map(node => (
                    <div key={node.data.id} className="bg-gray-800/50 p-2 rounded border border-gray-700/50">
                      <div className="flex justify-between items-center mb-1">
                        <span className="text-[10px] font-mono text-amber-400 truncate w-24">{node.data.id}</span>
                        <span className="text-[10px] text-gray-400">{trafficSettings[node.data.id] || 0} MPS</span>
                      </div>
                      <input
                        type="range" min="0" max="500" step="10"
                        value={trafficSettings[node.data.id] || 0}
                        onChange={(e) => handleTrafficChange(node.data.id, e.target.value)}
                        className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-amber-600"
                      />
                    </div>
                  ))}
                </div>

                {/* Siege List */}
                <div className="flex-1 overflow-auto p-3 space-y-3">
                  <div className="mb-2">
                    <label className="text-[10px] text-gray-500 uppercase block mb-1">Weather conditions</label>
                    <select
                        onChange={(e) => setWeather(e.target.value)}
                        className="w-full bg-gray-800 border border-gray-700 text-[10px] rounded p-1 outline-none text-gray-300 mb-2"
                        value={weather}
                    >
                        <option value="clear">Clear Skies</option>
                        <option value="storm">Thunderstorm</option>
                        <option value="fog">Thick Fog</option>
                        <option value="blizzard">Great Blizzard</option>
                    </select>

                    <label className="text-[10px] text-gray-500 uppercase block mb-1">Order Assault</label>
                    <div className="flex space-x-1 mb-1">
                      <select
                          id="siege-target"
                          className="flex-1 bg-gray-800 border border-gray-700 text-[10px] rounded p-1 outline-none text-gray-300"
                      >
                          <option value="">Target...</option>
                          {graphData.filter(el => el.data.id).map(node => (
                              <option key={node.data.id} value={node.data.id}>{node.data.id}</option>
                          ))}
                      </select>
                      <select
                          id="siege-type"
                          className="flex-1 bg-gray-800 border border-gray-700 text-[10px] rounded p-1 outline-none text-gray-300"
                      >
                          <option value="dragon_strike">Dragon</option>
                          <option value="plague">Plague</option>
                          <option value="famine">Famine</option>
                      </select>
                      <button
                        onClick={() => {
                          const target = document.getElementById('siege-target').value;
                          const type = document.getElementById('siege-type').value;
                          if (target) addFailure(target, type);
                        }}
                        className="bg-amber-700 p-1 rounded text-[10px]"
                      >
                        Go
                      </button>
                    </div>
                  </div>

                  {failures.map((f, i) => (
                    <div key={i} className="bg-orange-900/20 p-2 rounded border border-orange-500/30">
                        <div className="flex justify-between items-start mb-1">
                            <span className="text-[10px] font-bold text-orange-400 uppercase">{f.type.replace('_', ' ')}</span>
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

        {/* Middle Pane: War Map & Timeline */}
        <div className="flex-1 flex flex-col relative bg-gray-950">
          <div className="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700">
            <div className="flex items-center">
              <Castle size={16} className="mr-2 text-amber-500" />
              <span className="text-sm font-semibold text-gray-300 uppercase tracking-widest">The War Map</span>
            </div>
            {isSimulating && (
                <div className="flex items-center space-x-2 text-xs font-mono text-amber-400">
                    <Clock size={14} />
                    <span>Campaign Day {currentTime}</span>
                </div>
            )}
          </div>
          <div className="flex-1 bg-[url('https://www.transparenttextures.com/patterns/parchment.png')] bg-repeat">
            <CytoscapeComponent
              elements={[]}
              style={{ width: '100%', height: '100%' }}
              stylesheet={style}
              cy={(cy) => { cyRef.current = cy }}
            />
          </div>

          {/* Timeline Slider */}
          {isSimulating && simulation && (
            <div className="absolute bottom-6 left-1/2 -translate-x-1/2 w-3/4 bg-gray-800/95 backdrop-blur-md p-4 rounded-xl border border-amber-500/30 shadow-2xl">
                <div className="flex items-center space-x-4">
                    <button onClick={() => setIsSimulating(!isSimulating)} className="text-amber-500 hover:text-amber-400">
                        <Sword size={24} />
                    </button>
                    <input
                        type="range" min="0" max={simulation.timeline.length - 1}
                        value={currentTime}
                        onChange={(e) => setCurrentTime(parseInt(e.target.value))}
                        className="flex-1 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-amber-600"
                    />
                </div>
                <div className="flex justify-between mt-2 text-[10px] text-amber-500/50 font-mono uppercase tracking-widest">
                    <span>Dawn</span>
                    <span>Campaign Progress (Day {currentTime} / {simulation.timeline.length - 1})</span>
                    <span>Dusk</span>
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

        {/* Right Pane: Stronghold Stats & Scrolls */}
        <div className="w-1/4 flex flex-col border-l border-gray-800 bg-gray-900">
          {mirror && mirror.drifts.length > 0 && (
             <div className="bg-amber-900/20 border-b border-amber-900/50 p-3">
               <div className="flex items-center text-amber-500 text-[10px] font-bold uppercase mb-2">
                 <AlertCircle size={12} className="mr-1" /> Mirror Realm Discrepancy
               </div>
               {mirror.drifts.map((d, i) => (
                 <div key={i} className="text-[9px] text-amber-200/70 bg-black/20 p-1 rounded mb-1 border border-amber-900/20">
                   <span className="font-bold">{d.strongholdId}</span>: {d.description}
                 </div>
               ))}
             </div>
          )}

           <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-800">
            <Shield size={16} className="mr-2 text-amber-500" />
            <span className="text-sm font-semibold text-gray-300 uppercase tracking-widest">Stronghold Stats</span>
          </div>
          <div className="p-4 flex-1 overflow-auto">
            {simulation && simulation.timeline[currentTime] ? (
                <div className="space-y-4">
                    {Object.entries(simulation.timeline[currentTime].nodes).map(([id, m]) => (
                        <div key={id} className={`p-3 rounded border shadow-inner ${m.status === 'down' ? 'bg-red-900/30 border-red-500/50' : 'bg-gray-900 border-gray-800/50'}`}>
                            <div className="flex justify-between items-center mb-2">
                                <h4 className={`text-xs font-bold truncate ${m.status === 'down' ? 'text-red-400' : 'text-amber-400'}`}>{id}</h4>
                                <span className={`text-[8px] px-1 rounded uppercase font-bold ${
                                    m.status === 'down' ? 'bg-red-500 text-white' :
                                    m.status === 'degraded' ? 'bg-amber-500 text-white' : 'bg-green-600 text-white'
                                }`}>{m.status === 'up' ? 'standing' : m.status}</span>
                            </div>
                            <div className="grid grid-cols-2 gap-2 text-[10px]">
                                <div>
                                    <div className="text-gray-500 uppercase tracking-tighter text-[8px]">Garrison (CPU)</div>
                                    <div className={`font-mono ${m.cpuUsage > 80 ? 'text-red-400' : 'text-green-400'}`}>{m.cpuUsage.toFixed(1)}%</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase tracking-tighter text-[8px]">Messenger Delay</div>
                                    <div className="font-mono text-gray-300">{m.latency.toFixed(1)}ms</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase tracking-tighter text-[8px]">Battalions</div>
                                    <div className="font-mono text-gray-300">{m.replicas}</div>
                                </div>
                                <div>
                                    <div className="text-gray-500 uppercase tracking-tighter text-[8px]">Casualty Rate</div>
                                    <div className={`font-mono ${m.errorRate > 0 ? 'text-red-400 font-bold' : 'text-gray-300'}`}>{m.errorRate.toFixed(1)}%</div>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center text-gray-600 mt-20">
                    <Skull size={48} className="mx-auto mb-4 opacity-10" />
                    <p className="text-sm font-serif italic">Launch a Campaign to see the state of your realm.</p>
                </div>
            )}
          </div>

          <div className="h-1/3 border-t border-gray-800 flex flex-col bg-gray-950">
            <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-800">
                <Scroll size={16} className="mr-2 text-amber-500" />
                <span className="text-sm font-semibold text-gray-300 uppercase tracking-widest">The Scribe's Chronicles</span>
            </div>
            <div className="flex-1 p-4 overflow-auto text-[10px] text-amber-100/60 font-serif italic space-y-2">
                {simulation ? simulation.scrolls.map((s, i) => (
                  <div key={i} className="border-b border-amber-900/10 pb-1">~ {s}</div>
                )) : (
                  <div className="text-gray-600">The Scribe awaits your campaign orders...</div>
                )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;
