import api from "@/utils/api";

interface LoginResponse {
  auth_token: string;
  refresh_token: string;
}

export const loginApi = async (email: string, password: string) => {
  const response = await api.post<LoginResponse>("/auth/login", {
    email,
    password,
  });
  const { auth_token, refresh_token } = response.data;
  localStorage.setItem("token", auth_token);
  localStorage.setItem("refreshToken", refresh_token);
};

export const registerApi = async (
  email: string,
  password: string,
  confirmPassword: string,
) => {
  const response = await api.post<{ message: string }>("/auth/register", {
    email,
    password,
    confirmPassword,
  });
  return response.data;
};

export const logout = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("refreshToken");
};
