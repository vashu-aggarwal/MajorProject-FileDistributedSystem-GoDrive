import { useState, useEffect } from "react";

export default function MetricsDisplay({ darkMode, metrics = {} }) {
  const [bestAlgorithm, setBestAlgorithm] = useState(null);
  const [worstAlgorithm, setWorstAlgorithm] = useState(null);

  useEffect(() => {


    if (Object.keys(metrics).length === 0) {
      setBestAlgorithm(null);
      setWorstAlgorithm(null);
      return;
    }

    const speeds = Object.entries(metrics).map(([algo, data]) => ({
      algo,
      speed: parseFloat(data.speed) || 0,
    }));

    if (speeds.length > 0) {
      const sorted = [...speeds].sort((a, b) => b.speed - a.speed);
      setBestAlgorithm(sorted[0]);
      setWorstAlgorithm(sorted[sorted.length - 1]);
    }
  }, [metrics]);

  if (Object.keys(metrics).length === 0) {
    return null;
  }

  return (
    <div
      className={`p-4 rounded-xl shadow ${darkMode ? "border border-neutral-600" : "border border-black"
        }`}
    >
      <h2 className="text-xl font-bold mb-4">
        📈 Algorithm Performance Metrics
      </h2>

      {/* Metrics Table */}
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm border-collapse">
          <thead>
            <tr
              className={`${darkMode
                ? "bg-neutral-800 border-b border-neutral-600"
                : "bg-gray-200 border-b border-gray-400"
                }`}
            >
              <th className="border p-2 text-left">Algorithm</th>
              <th className="border p-2 text-center">Speed (MB/s)</th>
              <th className="border p-2 text-center">Size (KB)</th>
              <th className="border p-2 text-center">Time (s)</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(metrics).map(([algo, data]) => (
              <tr
                key={algo}
                className={`${darkMode
                  ? "border-b border-neutral-600 hover:bg-neutral-800"
                  : "border-b border-gray-300 hover:bg-gray-100"
                  } transition-colors`}
              >
                <td className="border p-2 font-semibold capitalize">
                  {algo}
                  {bestAlgorithm?.algo === algo && " 🏆"}
                  {worstAlgorithm?.algo === algo && " 🐢"}
                </td>
                <td className="border p-2 text-center font-semibold text-blue-600">
                  {data.speed}
                </td>
                <td className="border p-2 text-center">{data.size}</td>
                <td className="border p-2 text-center">{data.time}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

    </div>
  );
}
