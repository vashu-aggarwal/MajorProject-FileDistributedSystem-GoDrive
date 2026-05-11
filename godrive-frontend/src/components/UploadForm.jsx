import { useState, useEffect, useRef, useCallback } from "react";
import { uploadFile } from "../api/client";
import { FiUploadCloud, FiX, FiFile, FiCheckCircle, FiAlertTriangle, FiInfo } from "react-icons/fi";

// ─── Toast notification shown on the CARD face ───────────────────────────────
function Toast({ msg, onDismiss }) {
  if (!msg) return null;
  const styles = {
    success: "bg-emerald-500/90 text-white border-emerald-400",
    error:   "bg-red-500/90   text-white border-red-400",
    warning: "bg-amber-500/90 text-white border-amber-400",
    info:    "bg-blue-500/90  text-white border-blue-400",
  };
  const icons = {
    success: <FiCheckCircle  size={16} />,
    error:   <FiAlertTriangle size={16} />,
    warning: <FiAlertTriangle size={16} />,
    info:    <FiInfo          size={16} />,
  };
  return (
    <div
      className={`absolute bottom-4 left-4 right-4 z-50 flex items-start gap-2 px-4 py-3 rounded-2xl border shadow-xl backdrop-blur-sm text-sm animate-slideUp ${styles[msg.type] ?? styles.info}`}
    >
      <span className="mt-0.5 shrink-0">{icons[msg.type]}</span>
      <span className="flex-1 font-medium leading-snug">{msg.text}</span>
      <button onClick={onDismiss} className="ml-2 opacity-70 hover:opacity-100 transition shrink-0">
        <FiX size={14} />
      </button>
    </div>
  );
}

export default function UploadForm({ darkMode }) {
  const [showPopup,      setShowPopup]      = useState(false);
  const [fileName,       setFileName]       = useState("");
  const [content,        setContent]        = useState("");
  const [pickedFile,     setPickedFile]     = useState(null); // File object
  const [inputMode,      setInputMode]      = useState("text"); // "text" | "file"
  const [isDragging,     setIsDragging]     = useState(false);
  const [isLoading,      setIsLoading]      = useState(false);
  const [toast,          setToast]          = useState(null);
  const [uploadMetrics,  setUploadMetrics]  = useState(null);
  // duplicate-overwrite confirmation
  const [dupPending,     setDupPending]     = useState(false);

  const fileInputRef = useRef(null);
  const toastTimerRef = useRef(null);

  const showToast = useCallback((type, text, duration = 5000) => {
    clearTimeout(toastTimerRef.current);
    setToast({ type, text });
    if (duration > 0) {
      toastTimerRef.current = setTimeout(() => setToast(null), duration);
    }
  }, []);

  useEffect(() => () => clearTimeout(toastTimerRef.current), []);

  // Reset popup state on close
  const closePopup = () => {
    setShowPopup(false);
    setFileName("");
    setContent("");
    setPickedFile(null);
    setIsDragging(false);
    setDupPending(false);
  };

  // ── File picker / drag handlers ────────────────────────────────────────────
  const applyFile = (file) => {
    setPickedFile(file);
    setFileName(file.name);
    setInputMode("file");
  };

  const onFileInputChange = (e) => {
    const file = e.target.files?.[0];
    if (file) applyFile(file);
  };

  const onDrop = useCallback((e) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) applyFile(file);
  }, []);

  const onDragOver  = (e) => { e.preventDefault(); setIsDragging(true);  };
  const onDragLeave = ()  => setIsDragging(false);

  // ── Core upload logic ──────────────────────────────────────────────────────
  const doUpload = async () => {
    if (!fileName.trim()) {
      showToast("error", "Please enter a filename.");
      return;
    }
    if (inputMode === "text" && !content.trim()) {
      showToast("error", "Please enter file content or pick a file.");
      return;
    }

    setIsLoading(true);
    setDupPending(false);

    const payload   = inputMode === "file" ? pickedFile : content;
    const startTime = performance.now();
    const fileSize  = inputMode === "file"
      ? pickedFile.size
      : new Blob([content]).size;

    const result = await uploadFile(fileName.trim(), payload);
    const elapsed = (performance.now() - startTime) / 1000;

    setIsLoading(false);

    if (result.success) {
      const speedMBs   = fileSize / 1024 / 1024 / elapsed;
      setUploadMetrics({
        name:  fileName.trim(),
        size:  (fileSize / 1024).toFixed(2),
        time:  elapsed.toFixed(2),
        speed: isNaN(speedMBs) ? "0.00" : (speedMBs * 1024).toFixed(2), // KB/s
      });
      closePopup();
      showToast("success", `✅ "${fileName.trim()}" uploaded successfully!`);
    } else {
      // HTTP 409 Conflict → duplicate
      const errText = result.error || "";
      if (errText.includes("already present") || errText.includes("409") || errText.includes("Conflict")) {
        setIsLoading(false);
        setDupPending(true);
        showToast("warning", `⚠️ "${fileName.trim()}" already exists. Use Update to overwrite it.`, 0);
      } else {
        showToast("error", `Upload failed: ${errText || "Unknown error"}`);
      }
    }
  };

  const handleUpload = () => {
    setToast(null);
    doUpload();
  };

  // ── UI ─────────────────────────────────────────────────────────────────────
  const inputCls = `w-full px-4 py-3 rounded-2xl border outline-none transition-all text-sm
    ${darkMode
      ? "bg-[#1e293b] border-gray-700 text-white placeholder-gray-400 focus:border-blue-500"
      : "bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-500 focus:border-blue-500"}`;

  return (
    <div
      className={`relative h-[360px] rounded-3xl border overflow-hidden shadow-xl flex items-center justify-center transition-all duration-300
        ${darkMode ? "bg-[#111827] border-gray-800" : "bg-white border-gray-200"}`}
    >
      {/* ── Card face ── */}
      {!showPopup && (
        <div className="flex flex-col items-center justify-center text-center px-6">
          <button onClick={() => setShowPopup(true)} className="group">
            <div className={`w-24 h-24 rounded-full flex items-center justify-center transition-all duration-300 shadow-lg group-hover:scale-110
              ${darkMode ? "bg-blue-600 hover:bg-blue-500" : "bg-blue-500 hover:bg-blue-600"}`}>
              <FiUploadCloud size={42} className="text-white" />
            </div>
          </button>
          <h2 className={`mt-6 text-2xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Upload File</h2>
          <p className={`mt-2 text-sm ${darkMode ? "text-gray-400" : "text-gray-600"}`}>
            Click the icon to add a new file
          </p>

          {/* Last upload metrics badge */}
          {uploadMetrics && (
            <div className={`mt-4 px-4 py-2 rounded-2xl text-xs flex items-center gap-3 border
              ${darkMode ? "bg-blue-900/30 border-blue-800 text-blue-300" : "bg-blue-50 border-blue-200 text-blue-700"}`}>
              <span className="font-semibold truncate max-w-[120px]">{uploadMetrics.name}</span>
              <span>{uploadMetrics.size} KB</span>
              <span>{uploadMetrics.time}s</span>
              <span>{uploadMetrics.speed} KB/s</span>
            </div>
          )}
        </div>
      )}

      {/* ── Popup form ── */}
      {showPopup && (
        <div className={`absolute inset-0 overflow-y-auto p-6 flex flex-col gap-3
          ${darkMode ? "bg-[#0f172a]" : "bg-white"}`}>

          {/* Header */}
          <div className="flex items-center justify-between mb-1">
            <h2 className={`text-xl font-bold ${darkMode ? "text-white" : "text-gray-900"}`}>Upload New File</h2>
            <button onClick={closePopup}
              className={`p-2 rounded-xl transition ${darkMode ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
              <FiX size={20} className={darkMode ? "text-gray-300" : "text-gray-600"} />
            </button>
          </div>

          {/* Mode toggle */}
          <div className={`flex rounded-xl overflow-hidden border text-sm font-medium
            ${darkMode ? "border-gray-700" : "border-gray-300"}`}>
            {["text", "file"].map((m) => (
              <button key={m} onClick={() => { setInputMode(m); setPickedFile(null); setContent(""); }}
                className={`flex-1 py-2 transition-all ${inputMode === m
                  ? "bg-blue-600 text-white"
                  : darkMode ? "bg-[#1e293b] text-gray-400 hover:text-white" : "bg-gray-100 text-gray-500 hover:text-gray-800"}`}>
                {m === "text" ? "✍️ Text" : "📁 File"}
              </button>
            ))}
          </div>

          {/* Filename */}
          <input type="text" placeholder="Enter filename (e.g. notes.txt)"
            value={fileName} onChange={(e) => setFileName(e.target.value)}
            disabled={isLoading || inputMode === "file"}
            className={inputCls} />

          {/* Content / file picker */}
          {inputMode === "text" ? (
            <textarea rows={4} placeholder="Write file content…"
              value={content} onChange={(e) => setContent(e.target.value)}
              disabled={isLoading}
              className={`${inputCls} resize-none`} />
          ) : (
            <div
              onDrop={onDrop} onDragOver={onDragOver} onDragLeave={onDragLeave}
              onClick={() => fileInputRef.current?.click()}
              className={`flex flex-col items-center justify-center gap-2 py-6 rounded-2xl border-2 border-dashed cursor-pointer transition-all text-sm
                ${isDragging
                  ? "border-blue-500 bg-blue-500/10"
                  : darkMode ? "border-gray-600 hover:border-blue-500 bg-[#1e293b]" : "border-gray-300 hover:border-blue-400 bg-gray-50"}`}>
              <input ref={fileInputRef} type="file" className="hidden" onChange={onFileInputChange} />
              {pickedFile ? (
                <>
                  <FiFile size={28} className="text-blue-500" />
                  <span className={`font-medium truncate max-w-xs ${darkMode ? "text-white" : "text-gray-800"}`}>{pickedFile.name}</span>
                  <span className={`text-xs ${darkMode ? "text-gray-400" : "text-gray-500"}`}>
                    {(pickedFile.size / 1024).toFixed(1)} KB
                  </span>
                </>
              ) : (
                <>
                  <FiUploadCloud size={28} className={darkMode ? "text-gray-400" : "text-gray-400"} />
                  <span className={darkMode ? "text-gray-300" : "text-gray-600"}>
                    Drop a file or <span className="text-blue-500 font-semibold">Browse</span>
                  </span>
                </>
              )}
            </div>
          )}

          {/* Duplicate warning banner */}
          {dupPending && (
            <div className={`flex items-start gap-2 px-4 py-3 rounded-2xl border text-sm
              ${darkMode ? "bg-amber-900/30 border-amber-700 text-amber-300" : "bg-amber-50 border-amber-300 text-amber-800"}`}>
              <FiAlertTriangle size={16} className="mt-0.5 shrink-0" />
              <span>This file already exists in the system. Use <strong>Update</strong> to replace its content.</span>
            </div>
          )}

          {/* Upload button */}
          <button onClick={handleUpload} disabled={isLoading}
            className={`w-full py-3 rounded-2xl font-semibold transition-all duration-300 shadow-lg text-white flex items-center justify-center gap-2
              ${isLoading ? "bg-blue-400 cursor-not-allowed" : "bg-blue-600 hover:bg-blue-700 active:scale-95"}`}>
            {isLoading ? (
              <>
                <span className="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
                Uploading…
              </>
            ) : "Upload File"}
          </button>
        </div>
      )}

      {/* Toast (always visible over both states) */}
      <Toast msg={toast} onDismiss={() => setToast(null)} />
    </div>
  );
}
