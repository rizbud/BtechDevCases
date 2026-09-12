import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransferModal } from "./TransferModal";
import { ToastContainer } from "@/components/ToastContainer";
import { transferApi } from "@/wallet/utils/api";

vi.mock("@/wallet/utils/api", () => ({
  transferApi: vi.fn(),
}));

function renderModal(onClose: () => void) {
  const queryClient = new QueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <TransferModal open onClose={onClose} />
      <ToastContainer />
    </QueryClientProvider>,
  );
}

describe("TransferModal", () => {
  beforeEach(() => {
    vi.mocked(transferApi).mockReset();
  });

  it("submits recipient, amount, and notes and closes on success", async () => {
    vi.mocked(transferApi).mockResolvedValue({
      message: "ok",
      id: "1",
      recipient_id: "2",
      recipient_email: "recipient@example.com",
      amount: 50,
      created_at: new Date().toISOString(),
    });
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(
      screen.getByPlaceholderText("recipient@example.com"),
      "recipient@example.com",
    );
    await user.type(screen.getByPlaceholderText("0.00"), "50");
    await user.type(screen.getByPlaceholderText("Add a note"), "for lunch");
    await user.click(screen.getByRole("button", { name: "Transfer" }));

    await waitFor(() => expect(onClose).toHaveBeenCalled());
    expect(transferApi).toHaveBeenCalledWith(
      "recipient@example.com",
      50,
      "for lunch",
    );
  });

  it("shows a validation error and does not call the API for an invalid recipient", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(
      screen.getByPlaceholderText("recipient@example.com"),
      "test@test",
    );
    await user.type(screen.getByPlaceholderText("0.00"), "50");
    await user.click(screen.getByRole("button", { name: "Transfer" }));

    expect(
      await screen.findByText("Invalid email format"),
    ).toBeInTheDocument();
    expect(transferApi).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("shows a server-side field error returned from the API", async () => {
    const { AxiosError } = await import("axios");
    vi.mocked(transferApi).mockRejectedValue(
      new AxiosError(
        "Request failed",
        "400",
        undefined,
        undefined,
        {
          status: 400,
          data: { error: { recipient: "Recipient not found" } },
        } as never,
      ),
    );
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(
      screen.getByPlaceholderText("recipient@example.com"),
      "missing@example.com",
    );
    await user.type(screen.getByPlaceholderText("0.00"), "50");
    await user.click(screen.getByRole("button", { name: "Transfer" }));

    expect(await screen.findByText("Recipient not found")).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("shows a rate-limit message with retry time on 429", async () => {
    const { AxiosError } = await import("axios");
    vi.mocked(transferApi).mockRejectedValue(
      new AxiosError("Request failed", "429", undefined, undefined, {
        status: 429,
        headers: { "retry-after": "42" },
        data: { message: "Too many requests, please slow down" },
      } as never),
    );
    const onClose = vi.fn();
    const user = userEvent.setup();
    renderModal(onClose);

    await user.type(
      screen.getByPlaceholderText("recipient@example.com"),
      "recipient@example.com",
    );
    await user.type(screen.getByPlaceholderText("0.00"), "50");
    await user.click(screen.getByRole("button", { name: "Transfer" }));

    expect(
      await screen.findByText("Too many transfers. Try again in 42s."),
    ).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });
});
