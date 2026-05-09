import { useState, useEffect } from "react";
import { downloadFile } from "../api/client";

export default function DownloadForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [fileContent, setFileContent] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }
  const [isLoading, setIsLoading] = useState(false);
  const [downloadMetrics, setDownloadMetrics] = useState(null); // { size, time, speed }

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleDownload = async () => {
    setMessage(null);
    setDownloadMetrics(null);

    if (!fileName) {
      setMessage({ type: "error", text: "Please enter a filename" });
      return;
    }

    setIsLoading(true);
    const startTime = performance.now();

    const result = await downloadFile(fileName);

    const endTime = performance.now();
    const timeTaken = (endTime - startTime) / 1000; // Convert to seconds

    setIsLoading(false);

    if (result.success) {
      const data =
        typeof result.data === "string" ? JSON.parse(result.data) : result.data;
      const content = data.content || "";

      // Calculate metrics
      const fileSize = new Blob([content]).size;
      const speed = (fileSize / 1024 / 1024 / timeTaken).toFixed(2);

      const metrics = {
        size: (fileSize / 1024).toFixed(2),
        time: timeTaken.toFixed(2),
        speed: isNaN(speed) ? "0.00" : speed,
      };

      setDownloadMetrics(metrics);

      setFileContent(content);
      setMessage({ type: "success", text: "File downloaded successfully." });

      // Record metrics with current algorithm
      const currentConfig = localStorage.getItem("currentAlgorithm");
      if (currentConfig) {
        try {
          const config = JSON.parse(currentConfig);
          const algoName = config.algorithm || "unknown";
          const algorithmMetrics = JSON.parse(
            localStorage.getItem("algorithmMetrics") || "{}",
          );

          // Store metrics for this algorithm
          algorithmMetrics[algoName] = metrics;
          localStorage.setItem(
            "algorithmMetrics",
            JSON.stringify(algorithmMetrics),
          );

          // Dispatch custom event to notify AlgorithmSelector
          window.dispatchEvent(new Event("metricsUpdated"));
        } catch (e) {
          console.log("Could not record metrics");
        }
      }
    } else {
      setFileContent("");
      setMessage({ type: "error", text: "Download failed: " + result.error });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const textareaClass = `border p-3 w-full mt-2 rounded resize-none font-mono text-sm whitespace-pre-wrap break-words ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const messageClass = message
    ? message.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-2"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-2"
    : "";

  return (
    <div
      className={`p-4 rounded-xl shadow mb-4 ${
        darkMode ? "border border-neutral-600" : "border border-black"
      }`}
    >
      <h2 className="text-xl font-bold mb-2">Download File</h2>
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
        disabled={isLoading}
      />
      <button
        onClick={handleDownload}
        disabled={isLoading}
        className={`${isLoading ? "bg-green-400 cursor-not-allowed" : "bg-green-500 hover:bg-green-600"} text-white px-4 py-2 rounded transition-colors`}
      >
        {isLoading ? "Downloading..." : "Download"}
      </button>

      {message && <div className={messageClass}>{message.text}</div>}

      {/* Download Metrics */}
      {downloadMetrics && (
        <div
          className={`p-3 rounded mt-2 border ${
            darkMode
              ? "bg-neutral-800 border-neutral-600"
              : "bg-green-50 border-green-200"
          }`}
        >
          <h3 className="font-semibold mb-2">📊 Download Metrics</h3>
          <div className="grid grid-cols-3 gap-2 text-sm">
            <div>
              <span className="font-semibold">Size:</span>{" "}
              {downloadMetrics.size} KB
            </div>
            <div>
              <span className="font-semibold">Time:</span>{" "}
              {downloadMetrics.time}s
            </div>
            <div>
              <span className="font-semibold">Speed:</span>{" "}
              {downloadMetrics.speed} MB/s
            </div>
          </div>
        </div>
      )}

      {fileContent !== "" && (
        <>
          <label
            htmlFor="fileContent"
            className={`block mt-4 mb-1 font-semibold ${
              darkMode ? "text-gray-300" : "text-gray-700"
            }`}
          >
            File Content
          </label>
          <textarea
            id="fileContent"
            className={textareaClass}
            rows="8"
            readOnly
            value={fileContent}
            placeholder="No content to display"
          />
        </>
      )}
    </div>
  );
}
