import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TopUpModal } from "./TopUpModal";
import { ToastContainer } from "@/components/ToastContainer";
import { topUpApi } from "@/wallet/utils/api";

vi.mock("@/wallet/utils/api", () => ({
  topUpApi: vi.fn(),
}));

function renderModal(onClose: () => void) {
  const queryClient = new QueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <TopUpModal open onClose={onClose} />
      <ToastContainer />
    </QueryClientProvider>,
  );
}

describe("TopUpModal", () => {
  beforeEach(() => {
    vi.mocked(topUpApi).mockReset();
  });

  it("submits the amount and notes and closes on success", async () => {
    vi.mocked(topUpApi).mockResolvedValue({
      message: "ok",
      id: "1",
      recipient_id: "u1",
      recipient_email: "user@example.com",
      amount: 100,
      created_at: new Date().toISOString(),
    });
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(screen.getByPlaceholderText("0.00"), "100");
    await user.type(screen.getByPlaceholderText("Add a note"), "salary");
    await user.click(screen.getByRole("button", { name: "Top Up" }));

    await waitFor(() => expect(onClose).toHaveBeenCalled());
    expect(topUpApi).toHaveBeenCalledWith(100, "salary");
  });

  it("shows a validation error and does not call the API for a non-positive amount", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(screen.getByPlaceholderText("0.00"), "0");
    await user.click(screen.getByRole("button", { name: "Top Up" }));

    expect(
      await screen.findByText("Amount must be greater than zero"),
    ).toBeInTheDocument();
    expect(topUpApi).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("shows a rate-limit message with retry time on 429", async () => {
    const { AxiosError } = await import("axios");
    vi.mocked(topUpApi).mockRejectedValue(
      new AxiosError("Request failed", "429", undefined, undefined, {
        status: 429,
        headers: { "retry-after": "42" },
        data: { message: "Too many requests, please slow down" },
      } as never),
    );
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(screen.getByPlaceholderText("0.00"), "100");
    await user.click(screen.getByRole("button", { name: "Top Up" }));

    expect(
      await screen.findByText("Too many top ups. Try again in 42s."),
    ).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("does not render when open is false", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <TopUpModal open={false} onClose={vi.fn()} />
      </QueryClientProvider>,
    );
    expect(screen.queryByPlaceholderText("0.00")).not.toBeInTheDocument();
  });
});
