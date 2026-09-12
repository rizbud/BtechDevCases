import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 60000,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const url = originalRequest.url || "";
    if (
      error.response &&
      error.response.status === 401 &&
      !url.includes("/auth/")
    ) {
      if (!originalRequest._retry) {
        originalRequest._retry = true;
        try {
          const newToken = await refreshTokenApi();
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          return Promise.reject(refreshError);
        }
      }
    }
    return Promise.reject(error);
  },
);

const refreshTokenApi = async () => {
  const refreshToken = localStorage.getItem("refreshToken");
  if (refreshToken) {
    try {
      const response = await api.post("/auth/refresh-token", {
        refresh_token: refreshToken,
      });
      const { auth_token, refresh_token: newRefreshToken } = response.data;
      localStorage.setItem("token", auth_token);
      localStorage.setItem("refreshToken", newRefreshToken);
      return auth_token;
    } catch (error) {
      localStorage.removeItem("token");
      localStorage.removeItem("refreshToken");
      window.location.href = "/login";
      throw error;
    }
  } else {
    localStorage.removeItem("token");
    localStorage.removeItem("refreshToken");
    window.location.href = "/login";
    throw new Error("No refresh token available");
  }
};

export default api;
