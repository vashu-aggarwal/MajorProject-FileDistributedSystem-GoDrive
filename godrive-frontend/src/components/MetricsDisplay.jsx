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
      className={`p-4 rounded-xl shadow ${
        darkMode ? "border border-neutral-600" : "border border-black"
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
              className={`${
                darkMode
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
                className={`${
                  darkMode
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

      {/* Performance Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {bestAlgorithm && (
          <div
            className={`p-3 rounded border-l-4 border-green-500 ${
              darkMode ? "bg-neutral-800" : "bg-green-50"
            }`}
          >
            <div className="text-sm font-semibold text-green-600">
              Best Performance
            </div>
            <div className="text-lg font-bold capitalize">
              {bestAlgorithm.algo}
            </div>
            <div className="text-sm">{bestAlgorithm.speed} MB/s</div>
          </div>
        )}

        {worstAlgorithm && (
          <div
            className={`p-3 rounded border-l-4 border-red-500 ${
              darkMode ? "bg-neutral-800" : "bg-red-50"
            }`}
          >
            <div className="text-sm font-semibold text-red-600">
              Slowest Performance
            </div>
            <div className="text-lg font-bold capitalize">
              {worstAlgorithm.algo}
            </div>
            <div className="text-sm">{worstAlgorithm.speed} MB/s</div>
          </div>
        )}

        {bestAlgorithm &&
          worstAlgorithm &&
          bestAlgorithm.algo !== worstAlgorithm.algo && (
            <div
              className={`p-3 rounded border-l-4 border-purple-500 md:col-span-2 ${
                darkMode ? "bg-neutral-800" : "bg-purple-50"
              }`}
            >
              <div className="text-sm font-semibold text-purple-600">
                Performance Difference
              </div>
              <div className="text-lg font-bold">
                {(bestAlgorithm.speed / worstAlgorithm.speed).toFixed(2)}x
                faster
              </div>
              <div className="text-sm">
                {bestAlgorithm.algo} is{" "}
                {(bestAlgorithm.speed - worstAlgorithm.speed).toFixed(2)} MB/s
                faster than {worstAlgorithm.algo}
              </div>
            </div>
          )}
      </div>

      {/* Recommendations */}
      <div
        className={`p-3 rounded mt-3 border-l-4 border-yellow-500 ${
          darkMode ? "bg-neutral-800" : "bg-yellow-50"
        }`}
      >
        <div className="text-sm font-semibold text-yellow-700">
          💡 Recommendation
        </div>
        <div className="text-sm mt-1">
          {bestAlgorithm && (
            <>
              Use{" "}
              <span className="font-bold capitalize">{bestAlgorithm.algo}</span>{" "}
              for optimal performance with your current file sizes and system
              configuration.
            </>
          )}
        </div>
      </div>
    </div>
  );
}
