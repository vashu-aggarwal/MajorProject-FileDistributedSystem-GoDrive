import { useState, useEffect, useRef, useCallback } from "react";
import { downloadFile } from "../api/client";
import { FiDownloadCloud, FiX, FiCheckCircle, FiAlertTriangle, FiInfo, FiSave } from "react-icons/fi";

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

export default function DownloadForm({ darkMode }) {
  const [showPopup,       setShowPopup]       = useState(false);
  const [fileName,        setFileName]        = useState("");
  const [fileContent,     setFileContent]     = useState("");
  const [isLoading,       setIsLoading]       = useState(false);
  const [toast,           setToast]           = useState(null);
  const [downloadMetrics, setDownloadMetrics] = useState(null);
  const timerRef = useRef(null);

  const showToast = useCallback((type, text, duration = 5000) => {
    clearTimeout(timerRef.current);
    setToast({ type, text });
    if (duration > 0) timerRef.current = setTimeout(() => setToast(null), duration);
  }, []);

  useEffect(() => () => clearTimeout(timerRef.current), []);

  const closePopup = () => { setShowPopup(false); setFileName(""); setFileContent(""); setDownloadMetrics(null); };

  const handleDownload = async () => {
    if (!fileName.trim()) { showToast("error", "Please enter a filename."); return; }
    setToast(null); setFileContent(""); setDownloadMetrics(null); setIsLoading(true);
    const t0 = performance.now();
    const result = await downloadFile(fileName.trim());
    const elapsed = (performance.now() - t0) / 1000;
    setIsLoading(false);
    if (result.success) {
      const data    = typeof result.data === "string" ? JSON.parse(result.data) : result.data;
      const content = data.content || "";
      const size    = new Blob([content]).size;
      const spd     = size / 1024 / elapsed;
      setDownloadMetrics({ size: (size/1024).toFixed(2), time: elapsed.toFixed(2), speed: isNaN(spd) ? "0.00" : spd.toFixed(2) });
      setFileContent(content);
      showToast("success", `✅ "${fileName.trim()}" downloaded successfully!`);
    } else {
      const err = result.error || "";
      if (err.includes("404") || err.toLowerCase().includes("no such file")) {
        showToast("error", `❌ File "${fileName.trim()}" not found in the system.`);
      } else {
        showToast("error", `Download failed: ${err || "Unknown error"}`);
      }
    }
  };

  const saveToLocal = () => {
    const blob = new Blob([fileContent], { type: "text/plain" });
    const url  = URL.createObjectURL(blob);
    const a    = Object.assign(document.createElement("a"), { href: url, download: fileName });
    a.click(); URL.revokeObjectURL(url);
    showToast("info", "File saved to your device.");
  };

  const inp = `w-full px-4 py-3 rounded-2xl border outline-none transition-all text-sm
    ${darkMode ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-green-500"
               : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-green-500"}`;

  return (
    <div className={`relative h-[360px] rounded-3xl border overflow-hidden shadow-xl flex items-center justify-center transition-all duration-300
      ${darkMode ? "bg-[#111827] border-gray-800" : "bg-white border-gray-200"}`}>

      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button onClick={() => setShowPopup(true)} className="group">
            <div className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110
              ${darkMode ? "bg-green-600 hover:bg-green-500" : "bg-green-500 hover:bg-green-600"}`}>
              <FiDownloadCloud size={42} className="text-white"/>
            </div>
          </button>
          <h2 className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Download File</h2>
          <p className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"}`}>Click the icon to retrieve a file</p>
        </div>
      )}

      {showPopup && (
        <div className={`absolute inset-0 overflow-y-auto p-6 flex flex-col gap-3 ${darkMode ? "bg-[#0f172a]" : "bg-white"}`}>
          <div className="flex items-center justify-between mb-1">
            <h2 className={`text-xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Download File</h2>
            <button onClick={closePopup} className={`p-2 rounded-xl transition ${darkMode ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
              <FiX size={20} className={darkMode ? "text-gray-300" : "text-gray-600"}/>
            </button>
          </div>

          <input type="text" placeholder="Enter filename (e.g. report.txt)"
            value={fileName} onChange={(e) => setFileName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleDownload()}
            disabled={isLoading} className={inp}/>

          <button onClick={handleDownload} disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg text-white flex items-center justify-center gap-2
              ${isLoading ? "bg-green-400 cursor-not-allowed" : "bg-green-600 hover:bg-green-700 active:scale-95"}`}>
            {isLoading
              ? <><span className="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin"/>Downloading…</>
              : <><FiDownloadCloud size={16}/>Download File</>}
          </button>

          {downloadMetrics && (
            <div className={`p-3 rounded-2xl border text-sm grid grid-cols-3 gap-3
              ${darkMode ? "bg-[#1e293b] border-gray-700" : "bg-green-50 border-green-200"}`}>
              {[["Size",`${downloadMetrics.size} KB`],["Time",`${downloadMetrics.time}s`],["Speed",`${downloadMetrics.speed} KB/s`]].map(([l,v])=>(
                <div key={l} className="text-center">
                  <p className={`text-xs font-semibold mb-0.5 ${darkMode?"text-gray-400":"text-gray-500"}`}>{l}</p>
                  <p className={`font-bold ${darkMode?"text-white":"text-gray-800"}`}>{v}</p>
                </div>
              ))}
            </div>
          )}

          {fileContent && (
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <label className={`text-sm font-semibold ${darkMode?"text-gray-300":"text-gray-700"}`}>File Content</label>
                <button onClick={saveToLocal}
                  className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-xl font-medium transition
                    ${darkMode?"bg-green-700 hover:bg-green-600 text-white":"bg-green-100 hover:bg-green-200 text-green-800"}`}>
                  <FiSave size={12}/>Save to Device
                </button>
              </div>
              <textarea rows={5} readOnly value={fileContent}
                className={`w-full p-3 rounded-2xl border resize-none font-mono text-xs outline-none
                  ${darkMode?"bg-[#1e293b] border-gray-700 text-gray-100":"bg-gray-50 border-gray-300 text-gray-900"}`}/>
            </div>
          )}
        </div>
      )}

      <Toast msg={toast} onDismiss={() => setToast(null)}/>
    </div>
  );
}