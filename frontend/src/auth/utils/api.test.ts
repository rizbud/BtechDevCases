import { describe, expect, it, vi, beforeEach } from "vitest";
import api from "@/utils/api";
import { loginApi, logout, registerApi } from "./api";

vi.mock("@/utils/api", () => ({
  default: { post: vi.fn() },
}));

describe("loginApi", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(api.post).mockReset();
  });

  it("stores the auth and refresh tokens on success", async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { auth_token: "access-token", refresh_token: "refresh-token" },
    });

    await loginApi("user@example.com", "password123");

    expect(api.post).toHaveBeenCalledWith("/auth/login", {
      email: "user@example.com",
      password: "password123",
    });
    expect(localStorage.getItem("token")).toBe("access-token");
    expect(localStorage.getItem("refreshToken")).toBe("refresh-token");
  });
});

describe("registerApi", () => {
  beforeEach(() => {
    vi.mocked(api.post).mockReset();
  });

  it("posts email, password, and confirmPassword and returns the response", async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { message: "Registered successfully" },
    });

    const result = await registerApi(
      "user@example.com",
      "password123",
      "password123",
    );

    expect(api.post).toHaveBeenCalledWith("/auth/register", {
      email: "user@example.com",
      password: "password123",
      confirmPassword: "password123",
    });
    expect(result).toEqual({ message: "Registered successfully" });
  });
});

describe("logout", () => {
  it("clears the stored tokens", () => {
    localStorage.setItem("token", "a");
    localStorage.setItem("refreshToken", "b");

    logout();

    expect(localStorage.getItem("token")).toBeNull();
    expect(localStorage.getItem("refreshToken")).toBeNull();
  });
});
