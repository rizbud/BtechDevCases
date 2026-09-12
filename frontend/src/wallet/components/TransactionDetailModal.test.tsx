import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { TransactionDetailModal } from "./TransactionDetailModal";
import { useTransactionDetail } from "@/wallet/hooks/useTransactionDetail";

vi.mock("@/wallet/hooks/useTransactionDetail", () => ({
  useTransactionDetail: vi.fn(),
}));

describe("TransactionDetailModal", () => {
  beforeEach(() => {
    vi.mocked(useTransactionDetail).mockReset();
  });

  it("is closed when transactionId is null", () => {
    vi.mocked(useTransactionDetail).mockReturnValue({
      data: undefined,
      isLoading: false,
    } as ReturnType<typeof useTransactionDetail>);

    render(<TransactionDetailModal transactionId={null} onClose={vi.fn()} />);

    expect(screen.queryByText("Transaction Detail")).not.toBeInTheDocument();
  });

  it("shows a loading state while fetching", () => {
    vi.mocked(useTransactionDetail).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useTransactionDetail>);

    render(<TransactionDetailModal transactionId="tx-1" onClose={vi.fn()} />);

    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });

  it("renders transaction details, falling back to 'Top Up' when there is no sender", () => {
    vi.mocked(useTransactionDetail).mockReturnValue({
      data: {
        id: "tx-1",
        recipient_id: "u2",
        recipient_email: "recipient@example.com",
        amount: 50000,
        created_at: "2026-01-15T10:00:00Z",
      },
      isLoading: false,
    } as ReturnType<typeof useTransactionDetail>);

    render(<TransactionDetailModal transactionId="tx-1" onClose={vi.fn()} />);

    expect(screen.getByText("tx-1")).toBeInTheDocument();
    expect(screen.getByText("Top Up")).toBeInTheDocument();
    expect(screen.getByText("recipient@example.com")).toBeInTheDocument();
    expect(screen.getByText(/Rp\s?50\.000/)).toBeInTheDocument();
    expect(screen.getByText("-")).toBeInTheDocument();
  });

  it("shows the sender email and notes when present", () => {
    vi.mocked(useTransactionDetail).mockReturnValue({
      data: {
        id: "tx-2",
        sender_id: "u1",
        sender_email: "sender@example.com",
        recipient_id: "u2",
        recipient_email: "recipient@example.com",
        amount: 25000,
        notes: "for lunch",
        created_at: "2026-01-15T10:00:00Z",
      },
      isLoading: false,
    } as ReturnType<typeof useTransactionDetail>);

    render(<TransactionDetailModal transactionId="tx-2" onClose={vi.fn()} />);

    expect(screen.getByText("sender@example.com")).toBeInTheDocument();
    expect(screen.getByText("for lunch")).toBeInTheDocument();
  });
});
