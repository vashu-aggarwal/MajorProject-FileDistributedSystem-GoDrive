import { useState, useEffect } from "react";
import { getAlgorithms, setAlgorithms } from "../api/client";
import MetricsDisplay from "./MetricsDisplay";

export default function AlgorithmSelector({ darkMode }) {
  const [cacheAlgorithm, setCacheAlgorithm] = useState("lru");
  const [nodeSelectorAlgo, setNodeSelectorAlgo] = useState("roundRobin");
  const [currentConfig, setCurrentConfig] = useState(null);
  const [message, setMessage] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [metrics, setMetrics] = useState({}); // Store metrics for each algorithm

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  // Listen for metrics updates from upload/download operations
  useEffect(() => {
    const handleMetricsUpdate = () => {
      const savedMetrics = localStorage.getItem("algorithmMetrics");
      if (savedMetrics) {
        try {
          setMetrics(JSON.parse(savedMetrics));
        } catch (e) {
          console.log("Failed to update metrics");
        }
      }
    };

    window.addEventListener("metricsUpdated", handleMetricsUpdate);
    return () =>
      window.removeEventListener("metricsUpdated", handleMetricsUpdate);
  }, []);

  // Fetch current configuration on mount
  useEffect(() => {
    const fetchConfig = async () => {
      const result = await getAlgorithms();
      if (result.success) {
        const data =
          typeof result.data === "string"
            ? JSON.parse(result.data)
            : result.data;
        setCurrentConfig(data);
        setCacheAlgorithm(data.currentCache?.toLowerCase() || "lru");
        setNodeSelectorAlgo(data.currentSelector || "roundRobin");
      }
    };
    fetchConfig();

    // Load metrics from localStorage
    const savedMetrics = localStorage.getItem("algorithmMetrics");
    if (savedMetrics) {
      try {
        setMetrics(JSON.parse(savedMetrics));
      } catch (e) {
        console.log("Failed to load metrics from localStorage");
      }
    }
  }, []);

  const handleApply = async () => {
    setMessage(null);
    setIsLoading(true);
    const result = await setAlgorithms(cacheAlgorithm, nodeSelectorAlgo);
    setIsLoading(false);

    if (result.success) {
      const data =
        typeof result.data === "string" ? JSON.parse(result.data) : result.data;
      setCurrentConfig(data);

      // Save current algorithm to localStorage for metrics tracking
      localStorage.setItem(
        "currentAlgorithm",
        JSON.stringify({ algorithm: cacheAlgorithm }),
      );

      setMessage({
        type: "success",
        text: `✅ Algorithms updated: Cache=${cacheAlgorithm.toUpperCase()}, Selector=${nodeSelectorAlgo}`,
      });
    } else {
      setMessage({ type: "error", text: "Failed to update: " + result.error });
    }
  };

  // Function to record operation metrics
  const recordMetric = (algorithm, metricsData) => {
    setMetrics((prev) => {
      const updated = {
        ...prev,
        [algorithm]: metricsData,
      };
      localStorage.setItem("algorithmMetrics", JSON.stringify(updated));
      return updated;
    });
  };

  const clearMetrics = () => {
    setMetrics({});
    localStorage.removeItem("algorithmMetrics");
  };

  const selectClass = `border p-2 w-full mb-3 rounded ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 border-neutral-600"
      : "bg-gray-200 text-gray-900 border-neutral-900"
  }`;

  const labelClass = `block font-semibold mb-1 ${
    darkMode ? "text-gray-300" : "text-gray-700"
  }`;

  const messageClass =
    message?.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-3"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-3";

  const statusClass = `p-3 rounded mt-3 text-sm ${
    darkMode
      ? "bg-neutral-800 border border-neutral-700 text-gray-300"
      : "bg-gray-100 border border-gray-300 text-gray-700"
  }`;

  return (
    <>
      <div
        className={`p-4 rounded-xl shadow mb-4 ${
          darkMode ? "border border-neutral-600" : "border border-black"
        }`}
      >
        <h2 className="text-xl font-bold mb-4">⚙️ Algorithm Configuration</h2>

        <div>
          <label className={labelClass}>Cache Algorithm</label>
          <select
            value={cacheAlgorithm}
            onChange={(e) => setCacheAlgorithm(e.target.value)}
            disabled={isLoading}
            className={selectClass}
          >
            <option value="lru">LRU (Least Recently Used)</option>
            <option value="fifo">FIFO (First In First Out)</option>
            <option value="lfu">LFU (Least Frequently Used)</option>
            <option value="arc">ARC (Adaptive Replacement Cache)</option>
          </select>
        </div>

        <div>
          <label className={labelClass}>Node Selector Algorithm</label>
          <select
            value={nodeSelectorAlgo}
            onChange={(e) => setNodeSelectorAlgo(e.target.value)}
            disabled={isLoading}
            className={selectClass}
          >
            <option value="roundRobin">Round Robin</option>
            <option value="leastNode">Least Node (Load Balanced)</option>
            <option value="random">Random</option>
            <option value="powerOfTwo">Power of Two</option>
          </select>
        </div>

        <button
          onClick={handleApply}
          disabled={isLoading}
          className={`w-full ${isLoading ? "bg-purple-400 cursor-not-allowed" : "bg-purple-500 hover:bg-purple-600"} text-white px-4 py-2 rounded transition-colors font-semibold`}
        >
          {isLoading ? "Applying..." : "Apply Configuration"}
        </button>

        {message && <div className={messageClass}>{message.text}</div>}

        {currentConfig && (
          <div className={statusClass}>
            <div className="font-semibold mb-2">Current Configuration:</div>
            <div>
              🔵 Cache:{" "}
              <span className="font-mono font-bold">
                {currentConfig.currentCache?.toUpperCase() || "N/A"}
              </span>
            </div>
            <div>
              🟢 Selector:{" "}
              <span className="font-mono font-bold">
                {currentConfig.currentSelector || "N/A"}
              </span>
            </div>
            <div className="text-xs mt-2 text-gray-500">
              Capacity: {currentConfig.cacheCapacity} MB
            </div>
          </div>
        )}

        {Object.keys(metrics).length > 0 && (
          <div className="mt-4">
            <button
              onClick={clearMetrics}
              className={`w-full px-3 py-1 rounded text-sm ${
                darkMode
                  ? "bg-red-900 text-red-100 hover:bg-red-800"
                  : "bg-red-100 text-red-700 hover:bg-red-200"
              }`}
            >
              Clear Metrics History
            </button>
          </div>
        )}
      </div>

      {/* Metrics Display */}
      <MetricsDisplay darkMode={darkMode} metrics={metrics} />
    </>
  );
}
