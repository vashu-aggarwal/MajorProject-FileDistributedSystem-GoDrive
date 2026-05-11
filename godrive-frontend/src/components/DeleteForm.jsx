import { useState, useEffect, useRef, useCallback } from "react";
import { deleteFile } from "../api/client";
import { FiTrash2, FiX, FiCheckCircle, FiAlertTriangle, FiInfo } from "react-icons/fi";

function Toast({ msg, onDismiss }) {
  if (!msg) return null;
  const styles = {
    success: "bg-emerald-500/90 text-white border-emerald-400",
    error:   "bg-red-500/90 text-white border-red-400",
    warning: "bg-amber-500/90 text-white border-amber-400",
    info:    "bg-blue-500/90 text-white border-blue-400",
  };
  const icons = { success: <FiCheckCircle size={16}/>, error: <FiAlertTriangle size={16}/>, warning: <FiAlertTriangle size={16}/>, info: <FiInfo size={16}/> };
  return (
    <div className={`absolute bottom-4 left-4 right-4 z-50 flex items-start gap-2 px-4 py-3 rounded-2xl border shadow-xl text-sm animate-slideUp ${styles[msg.type]??styles.info}`}>
      <span className="mt-0.5 shrink-0">{icons[msg.type]}</span>
      <span className="flex-1 font-medium leading-snug">{msg.text}</span>
      <button onClick={onDismiss} className="ml-2 opacity-70 hover:opacity-100 transition shrink-0"><FiX size={14}/></button>
    </div>
  );
}

export default function DeleteForm({ darkMode }) {
  const [showPopup,   setShowPopup]   = useState(false);
  const [fileName,    setFileName]    = useState("");
  const [isLoading,   setIsLoading]   = useState(false);
  const [toast,       setToast]       = useState(null);
  // Two-step confirmation: false → "confirm" → deleting
  const [confirmed,   setConfirmed]   = useState(false);
  const timerRef = useRef(null);

  const showToast = useCallback((type, text, duration = 5000) => {
    clearTimeout(timerRef.current);
    setToast({ type, text });
    if (duration > 0) timerRef.current = setTimeout(() => setToast(null), duration);
  }, []);

  useEffect(() => () => clearTimeout(timerRef.current), []);

  const closePopup = () => { setShowPopup(false); setFileName(""); setConfirmed(false); };

  const handleDeleteClick = () => {
    if (!fileName.trim()) { showToast("error", "Please enter a filename."); return; }
    setToast(null);
    setConfirmed(true);  // show confirmation step
  };

  const confirmDelete = async () => {
    setIsLoading(true);
    const result = await deleteFile(fileName.trim());
    setIsLoading(false);
    setConfirmed(false);
    if (result.success) {
      closePopup();
      showToast("success", `🗑️ "${fileName.trim()}" deleted from the system.`);
    } else {
      const err = result.error || "";
      if (err.includes("404") || err.toLowerCase().includes("no such file")) {
        showToast("error", `❌ File "${fileName.trim()}" not found.`);
      } else {
        showToast("error", `Delete failed: ${err || "Unknown error"}`);
      }
    }
  };

  const inp = `w-full px-4 py-3 rounded-2xl border outline-none transition-all text-sm
    ${darkMode ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-red-500"
               : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-red-500"}`;

  return (
    <div className={`relative h-[360px] rounded-3xl border overflow-hidden shadow-xl flex items-center justify-center transition-all duration-300
      ${darkMode ? "bg-[#111827] border-gray-800" : "bg-white border-gray-200"}`}>

      {/* Card face */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button onClick={() => setShowPopup(true)} className="group">
            <div className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110
              ${darkMode ? "bg-red-600 hover:bg-red-500" : "bg-red-500 hover:bg-red-600"}`}>
              <FiTrash2 size={40} className="text-white"/>
            </div>
          </button>
          <h2 className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Delete File</h2>
          <p className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"}`}>Click the icon to permanently remove a file</p>
        </div>
      )}

      {/* Popup */}
      {showPopup && (
        <div className={`absolute inset-0 p-6 flex flex-col gap-3 justify-center ${darkMode ? "bg-[#0f172a]" : "bg-white"}`}>
          <div className="flex items-center justify-between mb-1">
            <h2 className={`text-xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Delete File</h2>
            <button onClick={closePopup} className={`p-2 rounded-xl transition ${darkMode ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
              <FiX size={20} className={darkMode ? "text-gray-300" : "text-gray-600"}/>
            </button>
          </div>

          <input type="text" placeholder="Enter filename to delete"
            value={fileName} onChange={(e) => { setFileName(e.target.value); setConfirmed(false); }}
            disabled={isLoading} className={inp}/>

          {/* Danger notice */}
          <div className={`flex items-start gap-2 px-4 py-3 rounded-2xl border text-sm
            ${darkMode ? "bg-red-950/30 border-red-900 text-red-300" : "bg-red-50 border-red-200 text-red-700"}`}>
            <FiAlertTriangle size={15} className="mt-0.5 shrink-0"/>
            <span>This action permanently removes the file from all distributed nodes and <strong>cannot be undone</strong>.</span>
          </div>

          {/* Confirmation step */}
          {!confirmed ? (
            <button onClick={handleDeleteClick} disabled={isLoading}
              className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg text-white flex items-center justify-center gap-2
                ${isLoading ? "bg-red-400 cursor-not-allowed" : "bg-red-600 hover:bg-red-700 active:scale-95"}`}>
              <FiTrash2 size={16}/> Delete File
            </button>
          ) : (
            <div className="flex flex-col gap-2">
              <p className={`text-sm text-center font-medium ${darkMode ? "text-amber-300" : "text-amber-700"}`}>
                Are you sure you want to delete <strong>"{fileName}"</strong>?
              </p>
              <div className="flex gap-2">
                <button onClick={() => setConfirmed(false)}
                  className={`flex-1 py-2.5 rounded-2xl font-semibold text-sm transition border
                    ${darkMode ? "border-gray-600 text-gray-300 hover:bg-gray-800" : "border-gray-300 text-gray-600 hover:bg-gray-100"}`}>
                  Cancel
                </button>
                <button onClick={confirmDelete} disabled={isLoading}
                  className={`flex-1 py-2.5 rounded-2xl font-semibold text-sm text-white transition flex items-center justify-center gap-1.5
                    ${isLoading ? "bg-red-400 cursor-not-allowed" : "bg-red-600 hover:bg-red-700 active:scale-95"}`}>
                  {isLoading
                    ? <><span className="w-3.5 h-3.5 border-2 border-white/40 border-t-white rounded-full animate-spin"/>Deleting…</>
                    : "Yes, Delete"}
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      <Toast msg={toast} onDismiss={() => setToast(null)}/>
    </div>
  );
}