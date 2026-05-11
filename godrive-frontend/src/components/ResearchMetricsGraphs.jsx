import { useEffect, useState } from "react";
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { getPerformanceMetrics } from "../api/client";

export default function ResearchMetricsGraphs({ darkMode }) {
  const [currentRun, setCurrentRun] = useState(null);
  const [series, setSeries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [tick, setTick] = useState(0);

  const fetchMetrics = async () => {
    setLoading(true);
    setError("");
    const result = await getPerformanceMetrics();
    setLoading(false);

    if (!result.success) {
      setError(result.error || "Failed to fetch metrics");
      return;
    }

    const payload =
      typeof result.data === "string" ? JSON.parse(result.data) : result.data;
    const run = payload.currentRun;
    setCurrentRun(run || null);

    if (run?.runId) {
      setSeries((prev) => {
        // Add a point only when request count changes so chart reflects real activity.
        if (
          prev.length > 0 &&
          prev[prev.length - 1].totalRequests === run.totalRequests
        ) {
          return prev;
        }

        const nextPoint = {
          sample: prev.length + 1,
          totalRequests: run.totalRequests || 0,
          throughputReqSec: Number((run.throughputReqSec || 0).toFixed(2)),
        };

        const updated = [...prev, nextPoint];
        if (updated.length > 40) {
          return updated.slice(updated.length - 40);
        }
        return updated;
      });
    }
  };

  useEffect(() => {
    fetchMetrics();
    const timer = setInterval(() => {
      setTick((value) => value + 1);
      fetchMetrics();
    }, 2500);
    return () => clearInterval(timer);
  }, []);

  const panelClass = darkMode
    ? "bg-neutral-900 border border-neutral-700"
    : "bg-white border border-gray-300";

  const textClass = darkMode ? "text-gray-200" : "text-gray-800";

  const chartBg = darkMode ? "#171717" : "#ffffff";
  const axisColor = darkMode ? "#e5e7eb" : "#111827";

  return (
    <section
      className={`mt-10 p-4 rounded-xl shadow ${panelClass} ${textClass}`}
    >
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 mb-4">
        <h2 className="text-xl font-bold">Live Throughput Graph</h2>
        <button
          onClick={fetchMetrics}
          className="px-3 py-2 rounded bg-blue-500 hover:bg-blue-600 text-white"
        >
          Refresh
        </button>
      </div>

      {loading && <div className="text-sm mb-3">Loading metrics...</div>}
      {error && <div className="text-sm mb-3 text-red-500">{error}</div>}

      {currentRun && (
        <div className="text-sm mb-6 grid grid-cols-2 md:grid-cols-3 gap-3">
          <div>
            <div className="font-semibold">Algorithm Pair</div>
            <div>
              {currentRun.cacheAlgorithm} / {currentRun.nodeSelectorAlgo}
            </div>
          </div>
          <div>
            <div className="font-semibold">Requests</div>
            <div>{currentRun.totalRequests}</div>
          </div>
          <div>
            <div className="font-semibold">Throughput</div>
            <div>{(currentRun.throughputReqSec || 0).toFixed(2)} req/s</div>
          </div>
        </div>
      )}

      {series.length === 0 ? (
        <div className="text-sm">
          No data points yet. Choose algorithms, then run
          upload/update/delete/download operations.
        </div>
      ) : (
        <div className="rounded-lg p-3 border border-gray-500/30">
          <h3 className="font-semibold mb-2">
            X-axis: Sample, Y-axis: Throughput (req/s)
          </h3>
          <div className="h-80" style={{ background: chartBg }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={series}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="sample" stroke={axisColor} />
                <YAxis
                  stroke={axisColor}
                  label={{
                    value: "Throughput (req/s)",
                    angle: -90,
                    position: "insideLeft",
                    fill: axisColor,
                  }}
                />
                <Tooltip />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="throughputReqSec"
                  stroke="#0ea5e9"
                  name="Throughput"
                  strokeWidth={2}
                  dot={{ r: 3 }}
                  isAnimationActive={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      <div className="text-xs mt-3 opacity-80">Auto-refresh tick: {tick}</div>
    </section>
  );
}
