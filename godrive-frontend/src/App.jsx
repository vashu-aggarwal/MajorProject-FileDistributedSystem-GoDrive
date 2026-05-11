import { useState, useEffect } from "react";
import { IoSunnyOutline, IoMoonOutline } from "react-icons/io5";

import UploadForm from "./components/UploadForm";
import DownloadForm from "./components/DownloadForm";
import UpdateForm from "./components/UpdateForm";
import DeleteForm from "./components/DeleteForm";
import AlgorithmSelector from "./components/AlgorithmSelector";

export default function App() {
  const [darkMode, setDarkMode] = useState(true);

  useEffect(() => {
    const savedMode = localStorage.getItem("darkMode");
    if (savedMode !== null) {
      setDarkMode(savedMode === "true");
    }
  }, []);

  useEffect(() => {
    localStorage.setItem("darkMode", darkMode);
  }, [darkMode]);

  return (
    <div
      className={`min-h-screen transition-all duration-500 ${darkMode
          ? "bg-[#0f172a] text-gray-100"
          : "bg-[#f1f5f9] text-gray-900"
        }`}
    >
      {/* Top Navbar */}
      <header
        className={`sticky top-0 z-50 border-b backdrop-blur-xl ${darkMode
            ? "bg-[#111827]/80 border-gray-800"
            : "bg-white/80 border-gray-200"
          }`}
      >
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          {/* Logo Section */}
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              GoDrive
            </h1>

            <p
              className={`text-sm mt-1 ${darkMode ? "text-gray-400" : "text-gray-600"
                }`}
            >
              Distributed Secure File Storage Dashboard
            </p>
          </div>

          {/* Theme Toggle */}
          <button
            aria-label="Toggle Theme"
            onClick={() => setDarkMode(!darkMode)}
            className={`p-3 rounded-2xl transition-all duration-300 shadow-lg border ${darkMode
                ? "bg-gray-800 border-gray-700 hover:bg-gray-700"
                : "bg-white border-gray-300 hover:bg-gray-100"
              }`}
          >
            {darkMode ? (
              <IoMoonOutline
                size={22}
                className="text-yellow-300"
              />
            ) : (
              <IoSunnyOutline
                size={22}
                className="text-yellow-500"
              />
            )}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-10">
        {/* Welcome Banner */}
        <div
          className={`mb-10 rounded-3xl p-8 border shadow-xl ${darkMode
              ? "bg-gradient-to-r from-blue-900/40 to-slate-900 border-gray-800"
              : "bg-gradient-to-r from-blue-100 to-white border-gray-200"
            }`}
        >
          <h2 className="text-3xl font-bold mb-3">
            File Management Dashboard
          </h2>

          <p
            className={`max-w-3xl leading-relaxed ${darkMode ? "text-gray-300" : "text-gray-700"
              }`}
          >
            Securely upload, update, download, and manage files
            across the distributed GoDrive storage system with
            encryption algorithm support and real-time operations.
          </p>
        </div>

        {/* Algorithm Selector */}
        <div
          className={`rounded-3xl p-6 mb-10 border shadow-xl ${darkMode
              ? "bg-[#111827] border-gray-800"
              : "bg-white border-gray-200"
            }`}
        >
          <div className="mb-4">
            <h3 className="text-xl font-semibold">
              Encryption Configuration
            </h3>

            <p
              className={`text-sm mt-1 ${darkMode ? "text-gray-400" : "text-gray-600"
                }`}
            >
              Select the preferred encryption algorithm for secure
              file operations.
            </p>
          </div>

          <AlgorithmSelector darkMode={darkMode} />
        </div>

        {/* Operations Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {[UploadForm, UpdateForm, DeleteForm, DownloadForm].map(
            (FormComponent, index) => (
              <div
                key={index}
                className={`rounded-3xl p-6 border shadow-xl transition-all duration-300 hover:-translate-y-1 hover:shadow-2xl ${darkMode
                    ? "bg-[#111827] border-gray-800"
                    : "bg-white border-gray-200"
                  }`}
              >
                <FormComponent darkMode={darkMode} />
              </div>
            )
          )}
        </div>
      </main>

      {/* Footer */}
      <footer
        className={`mt-10 border-t ${darkMode
            ? "border-gray-800 bg-[#111827]"
            : "border-gray-200 bg-white"
          }`}
      >
        <div className="max-w-7xl mx-auto px-6 py-5 flex flex-col md:flex-row justify-between items-center">
          <p
            className={`text-sm ${darkMode ? "text-gray-400" : "text-gray-600"
              }`}
          >
            © 2025 GoDrive. All rights reserved.
          </p>

          <p
            className={`text-sm mt-2 md:mt-0 ${darkMode ? "text-gray-500" : "text-gray-500"
              }`}
          >
            Secure Distributed File Management System
          </p>
        </div>
      </footer>
    </div>
  );
}
