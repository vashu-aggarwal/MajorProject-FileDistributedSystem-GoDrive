import { useState, useEffect, useRef } from "react";
import { uploadFile } from "../api/client";

export default function UploadForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [content, setContent] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }
  const [isLoading, setIsLoading] = useState(false);
  const [uploadMetrics, setUploadMetrics] = useState(null); // { size, time, speed }
  const [uploadMode, setUploadMode] = useState("text"); // "text" or "file"
  const [selectedFile, setSelectedFile] = useState(null);
  const [filePreview, setFilePreview] = useState("");
  const fileInputRef = useRef(null);

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleFileSelect = async (event) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setSelectedFile(file);
    setFileName(file.name);

    // For text files, show preview
    if (
      file.type.startsWith("text/") ||
      file.name.endsWith(".txt") ||
      file.name.endsWith(".pdf") ||
      file.name.endsWith(".docx")
    ) {
      const reader = new FileReader();
      reader.onload = (e) => {
        if (file.type.startsWith("text/")) {
          setFilePreview(e.target.result);
        } else {
          setFilePreview(
            `[Binary file: ${file.name} (${(file.size / 1024).toFixed(2)} KB)]`,
          );
        }
      };
      reader.readAsText(file);
    }
  };

  const handleUpload = async () => {
    setMessage(null);
    setUploadMetrics(null);

    if (!fileName) {
      setMessage({
        type: "error",
        text: "Please provide a filename",
      });
      return;
    }

    // Validate content based on upload mode
    if (uploadMode === "text" && !content) {
      setMessage({
        type: "error",
        text: "Please fill in filename and content",
      });
      return;
    }

    if (uploadMode === "file" && !selectedFile) {
      setMessage({
        type: "error",
        text: "Please select a file",
      });
      return;
    }

    setIsLoading(true);
    const startTime = performance.now();
    let fileSize = 0;

    let result;
    if (uploadMode === "text") {
      fileSize = new Blob([content]).size;
      result = await uploadFile(fileName, content);
    } else {
      fileSize = selectedFile.size;
      result = await uploadFile(fileName, selectedFile);
    }

    const endTime = performance.now();
    const timeTaken = (endTime - startTime) / 1000; // Convert to seconds
    const speed = (fileSize / 1024 / 1024 / timeTaken).toFixed(2); // MB/s

    const metrics = {
      size: (fileSize / 1024).toFixed(2),
      time: timeTaken.toFixed(2),
      speed: isNaN(speed) ? "0.00" : speed,
    };

    setUploadMetrics(metrics);

    setIsLoading(false);

    if (result.success) {
      setMessage({ type: "success", text: result.data });

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

      setFileName("");
      setContent("");
      setSelectedFile(null);
      setFilePreview("");
      if (fileInputRef.current) fileInputRef.current.value = "";
    } else {
      setMessage({ type: "error", text: "Upload failed: " + result.error });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const textareaClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 resize-none ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const messageClass =
    message?.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-2"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-2";

  const buttonClass = (color) =>
    `${
      isLoading
        ? `${color}-400 cursor-not-allowed`
        : `${color}-500 hover:${color}-600`
    } text-white px-4 py-2 rounded transition-colors`;

  return (
    <div
      className={`p-4 rounded-xl shadow mb-4 ${
        darkMode ? "border border-neutral-600" : "border border-black"
      }`}
    >
      <h2 className="text-xl font-bold mb-2">Upload File</h2>

      {/* Upload Mode Toggle */}
      <div className="flex gap-2 mb-4">
        <button
          onClick={() => {
            setUploadMode("text");
            setSelectedFile(null);
            setFilePreview("");
          }}
          disabled={isLoading}
          className={`px-4 py-2 rounded transition-colors ${
            uploadMode === "text"
              ? "bg-blue-500 text-white"
              : darkMode
                ? "bg-neutral-700 text-neutral-300 hover:bg-neutral-600"
                : "bg-gray-300 text-gray-900 hover:bg-gray-400"
          }`}
        >
          Text Mode
        </button>
        <button
          onClick={() => {
            setUploadMode("file");
            setContent("");
          }}
          disabled={isLoading}
          className={`px-4 py-2 rounded transition-colors ${
            uploadMode === "file"
              ? "bg-blue-500 text-white"
              : darkMode
                ? "bg-neutral-700 text-neutral-300 hover:bg-neutral-600"
                : "bg-gray-300 text-gray-900 hover:bg-gray-400"
          }`}
        >
          File Mode
        </button>
      </div>

      {/* Filename Input */}
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
        disabled={isLoading}
      />

      {/* Text Mode */}
      {uploadMode === "text" && (
        <textarea
          className={textareaClass}
          rows="4"
          placeholder="File content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          disabled={isLoading}
        />
      )}

      {/* File Mode */}
      {uploadMode === "file" && (
        <>
          <input
            ref={fileInputRef}
            type="file"
            onChange={handleFileSelect}
            disabled={isLoading}
            className={`border p-2 w-full mb-2 rounded ${
              darkMode
                ? "bg-neutral-900 text-neutral-100 border-neutral-600"
                : "bg-gray-200 text-gray-900 border-neutral-900"
            }`}
            accept=".txt,.pdf,.docx,.xlsx,.doc,.csv,.json,text/*"
          />
          {selectedFile && (
            <div
              className={`p-2 rounded mb-2 ${
                darkMode
                  ? "bg-neutral-800 border border-neutral-600"
                  : "bg-gray-100 border border-gray-300"
              }`}
            >
              <p className="text-sm">
                <span className="font-semibold">File:</span> {selectedFile.name}
              </p>
              <p className="text-sm">
                <span className="font-semibold">Size:</span>{" "}
                {(selectedFile.size / 1024).toFixed(2)} KB
              </p>
            </div>
          )}
          {filePreview && (
            <textarea
              className={`${textareaClass} mb-2`}
              rows="3"
              readOnly
              value={filePreview}
              placeholder="File preview"
            />
          )}
        </>
      )}

      {/* Upload Button */}
      <button
        onClick={handleUpload}
        disabled={isLoading}
        className={`bg-blue-500 hover:bg-blue-600 ${buttonClass("bg-blue")} text-white px-4 py-2 rounded transition-colors`}
      >
        {isLoading ? "Uploading..." : "Upload"}
      </button>

      {message && <div className={messageClass}>{message.text}</div>}

      {/* Upload Metrics */}
      {uploadMetrics && (
        <div
          className={`p-3 rounded mt-2 border ${
            darkMode
              ? "bg-neutral-800 border-neutral-600"
              : "bg-blue-50 border-blue-200"
          }`}
        >
          <h3 className="font-semibold mb-2">📊 Upload Metrics</h3>
          <div className="grid grid-cols-3 gap-2 text-sm">
            <div>
              <span className="font-semibold">Size:</span> {uploadMetrics.size}{" "}
              KB
            </div>
            <div>
              <span className="font-semibold">Time:</span> {uploadMetrics.time}s
            </div>
            <div>
              <span className="font-semibold">Speed:</span>{" "}
              {uploadMetrics.speed} MB/s
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
