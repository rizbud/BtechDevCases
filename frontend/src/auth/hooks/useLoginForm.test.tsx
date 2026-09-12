import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError } from "axios";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useLoginForm } from "./useLoginForm";
import { loginApi } from "@/auth/utils/api";

vi.mock("@/auth/utils/api", () => ({
  loginApi: vi.fn(),
}));

const navigate = vi.fn();
vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => navigate,
}));

function LoginTestForm() {
  const { register, errors, isSubmitting, onSubmit } = useLoginForm();
  return (
    <form onSubmit={onSubmit}>
      <input placeholder="email" {...register("email")} />
      {errors.email && <p>{errors.email.message}</p>}
      <input placeholder="password" {...register("password")} />
      {errors.password && <p>{errors.password.message}</p>}
      {errors.root && <p>{errors.root.message}</p>}
      <button type="submit" disabled={isSubmitting}>
        Login
      </button>
    </form>
  );
}

function renderForm() {
  const queryClient = new QueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <LoginTestForm />
    </QueryClientProvider>,
  );
}

describe("useLoginForm", () => {
  beforeEach(() => {
    vi.mocked(loginApi).mockReset();
    navigate.mockReset();
  });

  it("logs in and navigates to / on success", async () => {
    vi.mocked(loginApi).mockResolvedValue(undefined);
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "password123");
    await user.click(screen.getByRole("button", { name: "Login" }));

    await waitFor(() =>
      expect(loginApi).toHaveBeenCalledWith("user@example.com", "password123"),
    );
    expect(navigate).toHaveBeenCalledWith({ to: "/" });
  });

  it("maps server field errors onto the form and does not navigate", async () => {
    vi.mocked(loginApi).mockRejectedValue(
      new AxiosError("Request failed", "401", undefined, undefined, {
        status: 401,
        data: { error: { password: "Incorrect password" } },
      } as never),
    );
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "wrongpassword");
    await user.click(screen.getByRole("button", { name: "Login" }));

    expect(await screen.findByText("Incorrect password")).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });

  it("shows a root error message for a generic server error", async () => {
    vi.mocked(loginApi).mockRejectedValue(
      new AxiosError("Request failed", "500", undefined, undefined, {
        status: 500,
        data: { message: "Server unavailable" },
      } as never),
    );
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "password123");
    await user.click(screen.getByRole("button", { name: "Login" }));

    expect(await screen.findByText("Server unavailable")).toBeInTheDocument();
  });
});
