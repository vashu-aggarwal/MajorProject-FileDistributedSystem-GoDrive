import { useState, useEffect } from "react";
import { updateFile } from "../api/client";
import { FiEdit3, FiX } from "react-icons/fi";

export default function UpdateForm({ darkMode }) {
  const [showPopup, setShowPopup] = useState(false);

  const [fileName, setFileName] = useState("");
  const [content, setContent] = useState("");

  const [message, setMessage] = useState(null);

  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!message) return;

    const timer = setTimeout(() => {
      setMessage(null);
    }, 4000);

    return () => clearTimeout(timer);
  }, [message]);

  const handleUpdate = async () => {
    setMessage(null);

    if (!fileName || !content) {
      setMessage({
        type: "error",
        text: "Please enter filename and content",
      });
      return;
    }

    setIsLoading(true);

    const result = await updateFile(fileName, content);

    setIsLoading(false);

    if (result.success) {
      setMessage({
        type: "success",
        text:
          typeof result.data === "string"
            ? result.data
            : result.data.message,
      });

      setFileName("");
      setContent("");

      setShowPopup(false);
    } else {
      setMessage({
        type: "error",
        text: "Update failed",
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
      {/* Main Update UI */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button
            onClick={() => setShowPopup(true)}
            className="group"
          >
            <div
              className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110 ${darkMode
                  ? "bg-yellow-500 hover:bg-yellow-400"
                  : "bg-yellow-500 hover:bg-yellow-600"
                }`}
            >
              <FiEdit3
                size={40}
                className="text-white"
              />
            </div>
          </button>

          <h2
            className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"
              }`}
          >
            Update File
          </h2>

          <p
            className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"
              }`}
          >
            Click the icon to modify an existing file
          </p>
        </div>
      )}

      {/* Popup Form */}
      {showPopup && (
        <div
          className={`absolute inset-0 p-6 flex flex-col justify-center ${darkMode
              ? "bg-[#0f172a]"
              : "bg-white"
            }`}
        >
          {/* Close */}
          <button
            onClick={() => setShowPopup(false)}
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

          <h2
            className={`text-2xl font-bold mb-6 ${darkMode ? "text-white" : "text-gray-900"
              }`}
          >
            Update Existing File
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
            className={`w-full px-4 py-3 rounded-2xl border mb-4 outline-none transition-all ${darkMode
                ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-yellow-500"
                : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-yellow-500"
              }`}
          />

          {/* Content */}
          <textarea
            rows="5"
            placeholder="Write updated content..."
            value={content}
            onChange={(e) =>
              setContent(e.target.value)
            }
            disabled={isLoading}
            className={`w-full px-4 py-3 rounded-2xl border mb-5 resize-none outline-none transition-all ${darkMode
                ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-yellow-500"
                : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-yellow-500"
              }`}
          />

          {/* Update Button */}
          <button
            onClick={handleUpdate}
            disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg ${isLoading
                ? "bg-yellow-400 cursor-not-allowed"
                : "bg-yellow-500 hover:bg-yellow-600"
              } text-white`}
          >
            {isLoading
              ? "Updating..."
              : "Update File"}
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
        </div>
      )}
    </div>
  );
}