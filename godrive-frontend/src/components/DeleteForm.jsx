import { useState, useEffect } from "react";
import { deleteFile } from "../api/client";

export default function DeleteForm({ darkMode }) {
  const [fileName, setFileName] = useState("");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text: string }
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(() => setMessage(null), 5000);
    return () => clearTimeout(timer);
  }, [message]);

  const handleDelete = async () => {
    setMessage(null);
    if (!fileName) {
      setMessage({ type: "error", text: "Please enter a filename" });
      return;
    }

    setIsLoading(true);
    const result = await deleteFile(fileName);
    setIsLoading(false);

    if (result.success) {
      setMessage({ type: "success", text: result.data });
      setFileName("");
    } else {
      setMessage({ type: "error", text: "Delete failed: " + result.error });
    }
  };

  const inputClass = `border p-2 w-full mb-2 rounded placeholder-opacity-70 ${
    darkMode
      ? "bg-neutral-900 text-neutral-100 placeholder-neutral-400 border-neutral-600"
      : "bg-gray-200 text-gray-900 placeholder-neutral-900 border-neutral-900"
  }`;

  const messageClass =
    message?.type === "success"
      ? "bg-lime-100 border border-lime-900 text-lime-700 px-4 py-2 rounded mt-2"
      : "bg-red-100 border border-red-400 text-red-700 px-4 py-2 rounded mt-2";

  return (
    <div
      className={`p-4 rounded-xl shadow mb-4 ${
        darkMode ? "border border-neutral-600" : "border border-black"
      }`}
    >
      <h2 className="text-xl font-bold mb-2">Delete File</h2>
      <input
        className={inputClass}
        placeholder="Filename"
        value={fileName}
        onChange={(e) => setFileName(e.target.value)}
        disabled={isLoading}
      />
      <button
        onClick={handleDelete}
        disabled={isLoading}
        className={`${isLoading ? "bg-red-400 cursor-not-allowed" : "bg-red-500 hover:bg-red-600"} text-white px-4 py-2 rounded transition-colors`}
      >
        {isLoading ? "Deleting..." : "Delete"}
      </button>

      {message && <div className={messageClass}>{message.text}</div>}
    </div>
  );
}
