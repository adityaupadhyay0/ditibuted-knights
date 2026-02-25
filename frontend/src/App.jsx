import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import CytoscapeComponent from 'react-cytoscapejs';
import { Play, Code, Layout, Layers, AlertCircle } from 'lucide-react';

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

  const cyRef = useRef(null);

  const analyzeCode = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post(`${API_URL}/analyze`, { terraformCode });
      const { graph, manifests } = response.data;

      const elements = [
        ...graph.nodes.map(n => ({ data: { id: n.id, label: n.label, type: n.type } })),
        ...graph.edges.map(e => ({ data: { id: e.id, source: e.source, target: e.target } }))
      ];

      setGraphData(elements);
      setManifests(manifests);
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

  const layout = { name: 'cose', animate: true };

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
        'width': '100px',
        'height': '40px',
        'shape': 'round-rectangle'
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
        'curve-style': 'bezier',
        'control-point-step-size': 40
      }
    }
  ];

  return (
    <div className="flex flex-col h-screen bg-gray-900 text-white font-sans">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center space-x-2">
          <Layers className="text-indigo-500" size={24} />
          <h1 className="text-xl font-bold">InfraTwin <span className="text-sm font-normal text-gray-400">v0.1.0</span></h1>
        </div>
        <button
          onClick={analyzeCode}
          disabled={loading}
          className="flex items-center space-x-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-800 px-4 py-2 rounded-md transition"
        >
          {loading ? <span>Analyzing...</span> : <><Play size={16} /><span>Run Simulation</span></>}
        </button>
      </header>

      {/* Main Content */}
      <main className="flex flex-1 overflow-hidden">
        {/* Left Pane: Code Editor */}
        <div className="w-1/3 flex flex-col border-r border-gray-700">
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

        {/* Middle Pane: Graph Visualization */}
        <div className="flex-1 flex flex-col relative">
          <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-700">
            <Layout size={16} className="mr-2 text-gray-400" />
            <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">Topology Graph</span>
          </div>
          <div className="flex-1 bg-gray-900">
            <CytoscapeComponent
              elements={[]} // Start empty, handled by useEffect
              style={{ width: '100%', height: '100%' }}
              stylesheet={style}
              cy={(cy) => { cyRef.current = cy }}
            />
          </div>
          {error && (
            <div className="absolute bottom-4 left-4 right-4 bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded-md flex items-center space-x-3">
              <AlertCircle size={20} />
              <span>{error}</span>
            </div>
          )}
        </div>

        {/* Right Pane: K8s Manifests */}
        <div className="w-1/4 flex flex-col border-l border-gray-700 bg-gray-800">
           <div className="flex items-center px-4 py-2 bg-gray-800 border-b border-gray-700">
            <Layers size={16} className="mr-2 text-gray-400" />
            <span className="text-sm font-semibold text-gray-300 uppercase tracking-wider">K8s Manifests</span>
          </div>
          <pre className="flex-1 p-4 overflow-auto text-xs text-gray-400 font-mono">
            {manifests || "# No manifests generated yet."}
          </pre>
        </div>
      </main>
    </div>
  );
}

export default App;
