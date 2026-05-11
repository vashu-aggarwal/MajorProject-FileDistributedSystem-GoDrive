import { useState, useEffect } from "react";
import { deleteFile } from "../api/client";
import { FiTrash2, FiX } from "react-icons/fi";

export default function DeleteForm({ darkMode }) {
  const [showPopup, setShowPopup] = useState(false);

  const [fileName, setFileName] = useState("");

  const [message, setMessage] = useState(null);

  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!message) return;

    const timer = setTimeout(() => {
      setMessage(null);
    }, 4000);

    return () => clearTimeout(timer);
  }, [message]);

  const handleDelete = async () => {
    setMessage(null);

    if (!fileName) {
      setMessage({
        type: "error",
        text: "Please enter filename",
      });
      return;
    }

    setIsLoading(true);

    const result = await deleteFile(fileName);

    setIsLoading(false);

    if (result.success) {
      setMessage({
        type: "success",
        text: result.data,
      });

      setFileName("");

      setShowPopup(false);
    } else {
      setMessage({
        type: "error",
        text: "Delete failed",
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
      {/* Main Delete UI */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button
            onClick={() => setShowPopup(true)}
            className="group"
          >
            <div
              className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110 ${darkMode
                  ? "bg-red-600 hover:bg-red-500"
                  : "bg-red-500 hover:bg-red-600"
                }`}
            >
              <FiTrash2
                size={40}
                className="text-white"
              />
            </div>
          </button>

          <h2
            className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"
              }`}
          >
            Delete File
          </h2>

          <p
            className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"
              }`}
          >
            Click the icon to permanently remove a file
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
            className={`text-2xl font-bold mb-3 ${darkMode ? "text-white" : "text-gray-900"
              }`}
          >
            Delete File
          </h2>

          <p
            className={`text-sm mb-6 ${darkMode ? "text-gray-400" : "text-gray-600"
              }`}
          >
            This action cannot be undone.
          </p>

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
                ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-red-500"
                : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-red-500"
              }`}
          />

          {/* Warning Box */}
          <div
            className={`mb-5 p-4 rounded-2xl border text-sm ${darkMode
                ? "bg-red-950/30 border-red-900 text-red-300"
                : "bg-red-50 border-red-200 text-red-700"
              }`}
          >
            Warning: Deleting a file will permanently remove
            it from the distributed storage system.
          </div>

          {/* Delete Button */}
          <button
            onClick={handleDelete}
            disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg ${isLoading
                ? "bg-red-400 cursor-not-allowed"
                : "bg-red-600 hover:bg-red-700"
              } text-white`}
          >
            {isLoading
              ? "Deleting..."
              : "Delete File"}
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