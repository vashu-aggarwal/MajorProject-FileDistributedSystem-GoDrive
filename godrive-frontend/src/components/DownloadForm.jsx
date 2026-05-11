import { useState, useEffect } from "react";
import { downloadFile } from "../api/client";
import {
  FiDownloadCloud,
  FiX,
} from "react-icons/fi";

export default function DownloadForm({ darkMode }) {
  const [showPopup, setShowPopup] = useState(false);

  const [fileName, setFileName] = useState("");

  const [fileContent, setFileContent] =
    useState("");

  const [message, setMessage] = useState(null);

  const [isLoading, setIsLoading] =
    useState(false);

  const [downloadMetrics, setDownloadMetrics] =
    useState(null);

  useEffect(() => {
    if (!message) return;

    const timer = setTimeout(() => {
      setMessage(null);
    }, 4000);

    return () => clearTimeout(timer);
  }, [message]);

  const handleDownload = async () => {
    setMessage(null);

    setDownloadMetrics(null);

    if (!fileName) {
      setMessage({
        type: "error",
        text: "Please enter filename",
      });

      return;
    }

    setIsLoading(true);

    const startTime = performance.now();

    const result = await downloadFile(fileName);

    const endTime = performance.now();

    const timeTaken =
      (endTime - startTime) / 1000;

    setIsLoading(false);

    if (result.success) {
      const data =
        typeof result.data === "string"
          ? JSON.parse(result.data)
          : result.data;

      const content = data.content || "";

      const fileSize = new Blob([
        content,
      ]).size;

      const speed = (
        fileSize /
        1024 /
        1024 /
        timeTaken
      ).toFixed(2);

      const metrics = {
        size: (
          fileSize / 1024
        ).toFixed(2),

        time: timeTaken.toFixed(2),

        speed: isNaN(speed)
          ? "0.00"
          : speed,
      };

      setDownloadMetrics(metrics);

      setFileContent(content);

      setMessage({
        type: "success",
        text: "File downloaded successfully",
      });
    } else {
      setFileContent("");

      setMessage({
        type: "error",
        text: "Download failed",
      });
    }
  };

  return (
    <div
      className={`relative h-[360px] rounded-3xl border overflow-hidden shadow-xl flex items-center justify-center transition-all duration-300 ${darkMode
          ? "bg-[#111827] border-gray-800"
          : "bg-white border-gray-200"
        }`}
    >
      {/* Main Download UI */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button
            onClick={() => setShowPopup(true)}
            className="group"
          >
            <div
              className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110 ${darkMode
                  ? "bg-green-600 hover:bg-green-500"
                  : "bg-green-500 hover:bg-green-600"
                }`}
            >
              <FiDownloadCloud
                size={42}
                className="text-white"
              />
            </div>
          </button>

          <h2
            className={`mt-6 text-2xl font-bold ${darkMode
                ? "text-white"
                : "text-gray-900"
              }`}
          >
            Download File
          </h2>

          <p
            className={`mt-2 text-sm ${darkMode
                ? "text-gray-400"
                : "text-gray-600"
              }`}
          >
            Click the icon to retrieve a file
          </p>
        </div>
      )}

      {/* Popup */}
      {showPopup && (
        <div
          className={`absolute inset-0 overflow-y-auto p-6 ${darkMode
              ? "bg-[#0f172a]"
              : "bg-white"
            }`}
        >
          {/* Close */}
          <button
            onClick={() =>
              setShowPopup(false)
            }
            className={`absolute top-4 right-4 p-2 rounded-xl transition ${darkMode
                ? "hover:bg-gray-800"
                : "hover:bg-gray-100"
              }`}
          >
            <FiX
              size={22}
              className={
                darkMode
                  ? "text-gray-300"
                  : "text-gray-700"
              }
            />
          </button>

          {/* Heading */}
          <h2
            className={`text-2xl font-bold mb-6 ${darkMode
                ? "text-white"
                : "text-gray-900"
              }`}
          >
            Download File
          </h2>

          {/* Filename */}
          <input
            type="text"
            placeholder="Enter filename"
            value={fileName}
            onChange={(e) =>
              setFileName(e.target.value)
            }
            disabled={isLoading}
            className={`w-full px-4 py-3 rounded-2xl border mb-5 outline-none transition-all ${darkMode
                ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-green-500"
                : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-green-500"
              }`}
          />

          {/* Download Button */}
          <button
            onClick={handleDownload}
            disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg ${isLoading
                ? "bg-green-400 cursor-not-allowed"
                : "bg-green-600 hover:bg-green-700"
              } text-white`}
          >
            {isLoading
              ? "Downloading..."
              : "Download File"}
          </button>

          {/* Message */}
          {message && (
            <div
              className={`mt-4 px-4 py-3 rounded-2xl text-sm ${message.type === "success"
                  ? "bg-green-100 text-green-700 border border-green-300"
                  : "bg-red-100 text-red-700 border border-red-300"
                }`}
            >
              {message.text}
            </div>
          )}

          {/* Metrics */}
          {downloadMetrics && (
            <div
              className={`mt-5 p-4 rounded-2xl border ${darkMode
                  ? "bg-[#1e293b] border-gray-700"
                  : "bg-green-50 border-green-200"
                }`}
            >
              <h3 className="font-semibold mb-3">
                Download Metrics
              </h3>

              <div className="grid grid-cols-3 gap-4 text-sm">
                <div>
                  <p className="font-semibold">
                    Size
                  </p>

                  <p>
                    {downloadMetrics.size} KB
                  </p>
                </div>

                <div>
                  <p className="font-semibold">
                    Time
                  </p>

                  <p>
                    {downloadMetrics.time}s
                  </p>
                </div>

                <div>
                  <p className="font-semibold">
                    Speed
                  </p>

                  <p>
                    {(
                      downloadMetrics.speed *
                      1024
                    ).toFixed(2)}{" "}
                    KB/s
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* File Content */}
          {fileContent !== "" && (
            <div className="mt-5">
              <label
                className={`block mb-2 font-semibold ${darkMode
                    ? "text-gray-300"
                    : "text-gray-700"
                  }`}
              >
                File Content
              </label>

              <textarea
                rows="8"
                readOnly
                value={fileContent}
                className={`w-full p-4 rounded-2xl border resize-none font-mono text-sm ${darkMode
                    ? "bg-[#1e293b] border-gray-700 text-gray-100"
                    : "bg-gray-50 border-gray-300 text-gray-900"
                  }`}
              />
            </div>
          )}
        </div>
      )}
    </div>
  );
}