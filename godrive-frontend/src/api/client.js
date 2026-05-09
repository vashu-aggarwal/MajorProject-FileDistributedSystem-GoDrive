// Centralized API client for GoDrive backend
const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:4000";

/**
 * Wrapper for API calls with consistent error handling
 */
const apiCall = async (endpoint, options = {}) => {
  const url = `${API_BASE_URL}${endpoint}`;
  const defaultOptions = {
    headers: {
      "Content-Type": "application/json",
    },
    ...options,
  };

  try {
    const response = await fetch(url, defaultOptions);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }

    // Handle different content types
    const contentType = response.headers.get("content-type");
    if (contentType?.includes("application/json")) {
      return {
        success: true,
        data: await response.json(),
      };
    } else {
      return {
        success: true,
        data: await response.text(),
      };
    }
  } catch (error) {
    return {
      success: false,
      error: error.message,
    };
  }
};

/**
 * Upload a file to the backend
 * @param {string} fileName - The name of the file
 * @param {string | File} content - File content (string for text, File object for file uploads)
 * @returns {Promise} Upload result with success status
 */
export const uploadFile = async (fileName, content) => {
  // Handle File object (from file input)
  if (content instanceof File) {
    const formData = new FormData();
    formData.append("fileName", fileName);
    formData.append("file", content);

    try {
      const response = await fetch(`${API_BASE_URL}/upload`, {
        method: "POST",
        body: formData,
        // Don't set Content-Type header - let browser set it for FormData
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP ${response.status}`);
      }

      const contentType = response.headers.get("content-type");
      if (contentType?.includes("application/json")) {
        return {
          success: true,
          data: await response.json(),
        };
      } else {
        return {
          success: true,
          data: await response.text(),
        };
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
      };
    }
  }

  // Handle string content
  return apiCall("/upload", {
    method: "POST",
    body: JSON.stringify({ fileName, content }),
  });
};

/**
 * Download a file from the backend
 */
export const downloadFile = async (filename) => {
  return apiCall(`/download?filename=${encodeURIComponent(filename)}`, {
    method: "GET",
  });
};

/**
 * Update/replace an existing file
 */
export const updateFile = async (fileName, content) => {
  return apiCall("/update", {
    method: "PUT",
    body: JSON.stringify({ fileName, content }),
  });
};

/**
 * Delete a file from the backend
 */
export const deleteFile = async (filename) => {
  return apiCall(`/delete?filename=${encodeURIComponent(filename)}`, {
    method: "DELETE",
  });
};

/**
 * Get current algorithm configuration
 */
export const getAlgorithms = async () => {
  return apiCall("/config/algorithms", {
    method: "GET",
  });
};

/**
 * Update algorithm configuration
 */
export const setAlgorithms = async (cacheAlgorithm, nodeSelectorAlgo) => {
  return apiCall("/config/algorithms", {
    method: "POST",
    body: JSON.stringify({
      cacheAlgorithm,
      nodeSelectorAlgo,
      cacheCapacity: 100, // Default cache capacity in MB
    }),
  });
};

/**
 * Get cache status information
 */
export const getCacheStatus = async () => {
  return apiCall("/config/cache-status", {
    method: "GET",
  });
};

/**
 * Health check endpoint
 */
export const healthCheck = async () => {
  return apiCall("/", {
    method: "GET",
  });
};

/**
 * Get backend performance metrics for graph rendering
 */
export const getPerformanceMetrics = async () => {
  return apiCall("/metrics/performance", {
    method: "GET",
  });
};

/**
 * Reset backend metrics and start a new experimental run
 */
export const resetPerformanceMetrics = async (
  workloadId = "default",
  concurrency = 1,
) => {
  return apiCall("/metrics/reset", {
    method: "POST",
    body: JSON.stringify({ workloadId, concurrency }),
  });
};
