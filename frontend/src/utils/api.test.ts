import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { AxiosError, type AxiosAdapter, type InternalAxiosRequestConfig } from "axios";
import api from "./api";

const respond = (
  config: InternalAxiosRequestConfig,
  status: number,
  data: unknown,
) => {
  const response = { data, status, statusText: "", headers: {}, config };
  if (status >= 200 && status < 300) return response;
  throw new AxiosError(
    "Request failed",
    AxiosError.ERR_BAD_REQUEST,
    config,
    undefined,
    response,
  );
};

let originalHref = "";

beforeEach(() => {
  localStorage.clear();
  originalHref = window.location.href;
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { ...window.location, href: originalHref },
  });
});

afterEach(() => {
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { ...window.location, href: originalHref },
  });
});

describe("api request interceptor", () => {
  it("attaches the bearer token from localStorage when present", async () => {
    localStorage.setItem("token", "access-token");
    const adapter: AxiosAdapter = vi.fn(async (config) =>
      respond(config, 200, { ok: true }),
    );
    api.defaults.adapter = adapter;

    await api.get("/wallet/balance");

    expect(adapter).toHaveBeenCalledTimes(1);
    const config = vi.mocked(adapter).mock.calls[0][0];
    expect(config.headers.Authorization).toBe("Bearer access-token");
  });

  it("does not attach an Authorization header when no token is stored", async () => {
    const adapter: AxiosAdapter = vi.fn(async (config) =>
      respond(config, 200, { ok: true }),
    );
    api.defaults.adapter = adapter;

    await api.get("/wallet/balance");

    const config = vi.mocked(adapter).mock.calls[0][0];
    expect(config.headers.Authorization).toBeUndefined();
  });
});

describe("api response interceptor (401 refresh flow)", () => {
  it("refreshes the token and retries the original request on a 401", async () => {
    localStorage.setItem("token", "old-token");
    localStorage.setItem("refreshToken", "refresh-1");

    const adapter: AxiosAdapter = vi.fn(async (config) => {
      if (config.url === "/auth/refresh-token") {
        return respond(config, 200, {
          auth_token: "new-token",
          refresh_token: "refresh-2",
        });
      }
      if (config.url === "/wallet/balance") {
        if (config.headers.Authorization === "Bearer new-token") {
          return respond(config, 200, { balance: 100 });
        }
        return respond(config, 401, { message: "unauthorized" });
      }
      throw new Error(`unexpected url ${config.url}`);
    });
    api.defaults.adapter = adapter;

    const response = await api.get("/wallet/balance");

    expect(response.data).toEqual({ balance: 100 });
    expect(adapter).toHaveBeenCalledTimes(3);
    expect(localStorage.getItem("token")).toBe("new-token");
    expect(localStorage.getItem("refreshToken")).toBe("refresh-2");
  });

  it("clears tokens and redirects to /login when there is no refresh token", async () => {
    localStorage.setItem("token", "old-token");

    const adapter: AxiosAdapter = vi.fn(async (config) =>
      respond(config, 401, { message: "unauthorized" }),
    );
    api.defaults.adapter = adapter;

    await expect(api.get("/wallet/balance")).rejects.toBeTruthy();

    expect(localStorage.getItem("token")).toBeNull();
    expect(localStorage.getItem("refreshToken")).toBeNull();
    expect(window.location.href).toContain("/login");
  });

  it("clears tokens and redirects to /login when the refresh request itself fails", async () => {
    localStorage.setItem("token", "old-token");
    localStorage.setItem("refreshToken", "expired-refresh");

    const adapter: AxiosAdapter = vi.fn(async (config) => {
      if (config.url === "/auth/refresh-token") {
        return respond(config, 401, { message: "expired" });
      }
      return respond(config, 401, { message: "unauthorized" });
    });
    api.defaults.adapter = adapter;

    await expect(api.get("/wallet/balance")).rejects.toBeTruthy();

    expect(localStorage.getItem("token")).toBeNull();
    expect(localStorage.getItem("refreshToken")).toBeNull();
    expect(window.location.href).toContain("/login");
  });

  it("does not attempt a refresh for a 401 from an /auth/ endpoint", async () => {
    const adapter: AxiosAdapter = vi.fn(async (config) =>
      respond(config, 401, { message: "invalid credentials" }),
    );
    api.defaults.adapter = adapter;

    await expect(
      api.post("/auth/login", { email: "a@a.com", password: "wrong" }),
    ).rejects.toBeTruthy();

    expect(adapter).toHaveBeenCalledTimes(1);
  });
});
