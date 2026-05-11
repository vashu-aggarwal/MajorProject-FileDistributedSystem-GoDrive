import { useState, useEffect, useRef, useCallback } from "react";
import { updateFile } from "../api/client";
import { FiEdit3, FiX, FiCheckCircle, FiAlertTriangle, FiInfo, FiUploadCloud } from "react-icons/fi";

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

export default function UpdateForm({ darkMode }) {
  const [showPopup, setShowPopup] = useState(false);
  const [fileName,  setFileName]  = useState("");
  const [content,   setContent]   = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [toast,     setToast]     = useState(null);
  const [updateMetrics, setUpdateMetrics] = useState(null);
  const timerRef = useRef(null);

  const showToast = useCallback((type, text, duration = 5000) => {
    clearTimeout(timerRef.current);
    setToast({ type, text });
    if (duration > 0) timerRef.current = setTimeout(() => setToast(null), duration);
  }, []);

  useEffect(() => () => clearTimeout(timerRef.current), []);

  const closePopup = () => { setShowPopup(false); setFileName(""); setContent(""); };

  const handleUpdate = async () => {
    if (!fileName.trim()) { showToast("error", "Please enter a filename."); return; }
    if (!content.trim())  { showToast("error", "Please enter the new content."); return; }
    setToast(null); setIsLoading(true);
    const t0     = performance.now();
    const result = await updateFile(fileName.trim(), content);
    const elapsed = (performance.now() - t0) / 1000;
    setIsLoading(false);
    if (result.success) {
      const size = new Blob([content]).size;
      setUpdateMetrics({ name: fileName.trim(), size: (size/1024).toFixed(2), time: elapsed.toFixed(2) });
      closePopup();
      showToast("success", `✅ "${fileName.trim()}" updated successfully!`);
    } else {
      const err = result.error || "";
      if (err.includes("404") || err.toLowerCase().includes("no such file")) {
        showToast("error", `❌ File "${fileName.trim()}" not found. Upload it first.`);
      } else {
        showToast("error", `Update failed: ${err || "Unknown error"}`);
      }
    }
  };

  const inp = `w-full px-4 py-3 rounded-2xl border outline-none transition-all text-sm
    ${darkMode ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-yellow-500"
               : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-yellow-500"}`;

  return (
    <div className={`relative h-[360px] rounded-3xl border overflow-hidden shadow-xl flex items-center justify-center transition-all duration-300
      ${darkMode ? "bg-[#111827] border-gray-800" : "bg-white border-gray-200"}`}>

      {/* Card face */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button onClick={() => setShowPopup(true)} className="group">
            <div className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110
              ${darkMode ? "bg-yellow-500 hover:bg-yellow-400" : "bg-yellow-500 hover:bg-yellow-600"}`}>
              <FiEdit3 size={40} className="text-white"/>
            </div>
          </button>
          <h2 className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Update File</h2>
          <p className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"}`}>Click the icon to modify an existing file</p>

          {updateMetrics && (
            <div className={`mt-4 px-4 py-2 rounded-2xl text-xs flex items-center gap-3 border
              ${darkMode ? "bg-yellow-900/30 border-yellow-800 text-yellow-300" : "bg-yellow-50 border-yellow-200 text-yellow-800"}`}>
              <span className="font-semibold truncate max-w-[120px]">{updateMetrics.name}</span>
              <span>{updateMetrics.size} KB</span>
              <span>{updateMetrics.time}s</span>
            </div>
          )}
        </div>
      )}

      {/* Popup */}
      {showPopup && (
        <div className={`absolute inset-0 p-6 flex flex-col gap-3 justify-center ${darkMode ? "bg-[#0f172a]" : "bg-white"}`}>
          <div className="flex items-center justify-between mb-1">
            <h2 className={`text-xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Update Existing File</h2>
            <button onClick={closePopup} className={`p-2 rounded-xl transition ${darkMode ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
              <FiX size={20} className={darkMode ? "text-gray-300" : "text-gray-600"}/>
            </button>
          </div>

          <input type="text" placeholder="Enter filename to update (e.g. notes.txt)"
            value={fileName} onChange={(e) => setFileName(e.target.value)}
            disabled={isLoading} className={inp}/>

          <textarea rows={5} placeholder="Enter the new content to replace the file with…"
            value={content} onChange={(e) => setContent(e.target.value)}
            disabled={isLoading} className={`${inp} resize-none`}/>

          {/* Info notice */}
          <div className={`flex items-start gap-2 px-4 py-2.5 rounded-2xl border text-xs
            ${darkMode ? "bg-yellow-900/20 border-yellow-800 text-yellow-300" : "bg-yellow-50 border-yellow-200 text-yellow-800"}`}>
            <FiInfo size={13} className="mt-0.5 shrink-0"/>
            <span>This will replace the entire file content across all distributed nodes.</span>
          </div>

          <button onClick={handleUpdate} disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg text-white flex items-center justify-center gap-2
              ${isLoading ? "bg-yellow-400 cursor-not-allowed" : "bg-yellow-500 hover:bg-yellow-600 active:scale-95"}`}>
            {isLoading
              ? <><span className="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin"/>Updating…</>
              : <><FiUploadCloud size={16}/>Update File</>}
          </button>
        </div>
      )}

      <Toast msg={toast} onDismiss={() => setToast(null)}/>
    </div>
  );
}
