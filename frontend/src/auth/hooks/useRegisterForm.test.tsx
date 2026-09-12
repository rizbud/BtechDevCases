import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError } from "axios";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useRegisterForm } from "./useRegisterForm";
import { registerApi } from "@/auth/utils/api";

vi.mock("@/auth/utils/api", () => ({
  registerApi: vi.fn(),
}));

function RegisterTestForm() {
  const { register, errors, isSubmitting, isSuccess, onSubmit } =
    useRegisterForm();
  if (isSuccess) return <p>Registration complete!</p>;
  return (
    <form onSubmit={onSubmit}>
      <input placeholder="email" {...register("email")} />
      {errors.email && <p>{errors.email.message}</p>}
      <input placeholder="password" {...register("password")} />
      {errors.password && <p>{errors.password.message}</p>}
      <input placeholder="confirmPassword" {...register("confirmPassword")} />
      {errors.confirmPassword && <p>{errors.confirmPassword.message}</p>}
      {errors.root && <p>{errors.root.message}</p>}
      <button type="submit" disabled={isSubmitting}>
        Register
      </button>
    </form>
  );
}

function renderForm() {
  const queryClient = new QueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <RegisterTestForm />
    </QueryClientProvider>,
  );
}

describe("useRegisterForm", () => {
  beforeEach(() => {
    vi.mocked(registerApi).mockReset();
  });

  it("registers and shows the success state", async () => {
    vi.mocked(registerApi).mockResolvedValue({ message: "Registered" });
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "password123");
    await user.type(
      screen.getByPlaceholderText("confirmPassword"),
      "password123",
    );
    await user.click(screen.getByRole("button", { name: "Register" }));

    await waitFor(() =>
      expect(registerApi).toHaveBeenCalledWith(
        "user@example.com",
        "password123",
        "password123",
      ),
    );
    expect(await screen.findByText("Registration complete!")).toBeInTheDocument();
  });

  it("shows a client-side error when confirmPassword does not match", async () => {
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "password123");
    await user.type(
      screen.getByPlaceholderText("confirmPassword"),
      "different123",
    );
    await user.click(screen.getByRole("button", { name: "Register" }));

    expect(await screen.findByText("Passwords do not match")).toBeInTheDocument();
    expect(registerApi).not.toHaveBeenCalled();
  });

  it("maps a server field error (e.g. duplicate email) onto the form", async () => {
    vi.mocked(registerApi).mockRejectedValue(
      new AxiosError("Request failed", "409", undefined, undefined, {
        status: 409,
        data: { error: { email: "Email already registered" } },
      } as never),
    );
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByPlaceholderText("email"), "user@example.com");
    await user.type(screen.getByPlaceholderText("password"), "password123");
    await user.type(
      screen.getByPlaceholderText("confirmPassword"),
      "password123",
    );
    await user.click(screen.getByRole("button", { name: "Register" }));

    expect(
      await screen.findByText("Email already registered"),
    ).toBeInTheDocument();
  });
});
