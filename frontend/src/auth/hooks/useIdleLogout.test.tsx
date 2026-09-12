import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render } from "@testing-library/react";
import { act } from "@testing-library/react";
import { IDLE_TIMEOUT_MS, useIdleLogout } from "./useIdleLogout";
import { logout } from "@/auth/utils/api";

vi.mock("@/auth/utils/api", () => ({
  logout: vi.fn(),
}));

const navigate = vi.fn();
vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => navigate,
}));

function TestComponent() {
  useIdleLogout();
  return null;
}

describe("useIdleLogout", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    navigate.mockReset();
    vi.mocked(logout).mockReset();
    localStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("logs out and navigates to /login after 15 minutes of no activity", () => {
    localStorage.setItem("token", "access-token");
    render(<TestComponent />);

    act(() => {
      vi.advanceTimersByTime(IDLE_TIMEOUT_MS);
    });

    expect(logout).toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ to: "/login" });
  });

  it("does not log out if there is no token stored", () => {
    render(<TestComponent />);

    act(() => {
      vi.advanceTimersByTime(IDLE_TIMEOUT_MS);
    });

    expect(logout).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
  });

  it("resets the timer on activity, avoiding logout", () => {
    localStorage.setItem("token", "access-token");
    render(<TestComponent />);

    act(() => {
      vi.advanceTimersByTime(IDLE_TIMEOUT_MS - 1000);
    });
    act(() => {
      window.dispatchEvent(new Event("keydown"));
    });
    act(() => {
      vi.advanceTimersByTime(IDLE_TIMEOUT_MS - 1000);
    });

    expect(logout).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(logout).toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ to: "/login" });
  });
});
