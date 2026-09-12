import { describe, expect, it, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TransactionTable } from "./TransactionTable";
import { useTransactionHistory } from "@/wallet/hooks/useTransactionHistory";

vi.mock("@/wallet/hooks/useTransactionHistory", () => ({
  useTransactionHistory: vi.fn(),
}));

const baseHookReturn = {
  transactions: [],
  totalPages: 1,
  isLoading: false,
  isFetching: false,
  isError: false,
  error: null,
  refetch: vi.fn(),
  dateRange: { start_date: "2026-01-01", end_date: "2026-01-31" },
  setDateRange: vi.fn(),
  page: 1,
  setPage: vi.fn(),
};

describe("TransactionTable", () => {
  beforeEach(() => {
    vi.mocked(useTransactionHistory).mockReset();
  });

  it("shows a loading state", () => {
    vi.mocked(useTransactionHistory).mockReturnValue({
      ...baseHookReturn,
      isLoading: true,
    } as ReturnType<typeof useTransactionHistory>);

    const { container } = render(
      <TransactionTable onSelectTransaction={vi.fn()} />,
    );

    expect(container.querySelectorAll(".skeleton").length).toBeGreaterThan(0);
  });

  it("shows an empty state when there are no transactions", () => {
    vi.mocked(useTransactionHistory).mockReturnValue(
      baseHookReturn as ReturnType<typeof useTransactionHistory>,
    );

    render(<TransactionTable onSelectTransaction={vi.fn()} />);

    expect(
      screen.getByText("No transactions in this range."),
    ).toBeInTheDocument();
  });

  it("renders transaction rows and calls onSelectTransaction on click", async () => {
    vi.mocked(useTransactionHistory).mockReturnValue({
      ...baseHookReturn,
      transactions: [
        {
          id: "tx-1",
          sender_email: "sender@example.com",
          recipient_id: "u2",
          recipient_email: "recipient@example.com",
          amount: 25000,
          created_at: "2026-01-15T10:00:00Z",
        },
      ],
    } as ReturnType<typeof useTransactionHistory>);
    const onSelectTransaction = vi.fn();
    const user = userEvent.setup();
    render(<TransactionTable onSelectTransaction={onSelectTransaction} />);

    expect(screen.getByText("sender@example.com")).toBeInTheDocument();
    expect(screen.getByText("recipient@example.com")).toBeInTheDocument();

    await user.click(screen.getByText("recipient@example.com"));
    expect(onSelectTransaction).toHaveBeenCalledWith("tx-1");
  });

  it("falls back to 'Top Up' when a row has no sender", () => {
    vi.mocked(useTransactionHistory).mockReturnValue({
      ...baseHookReturn,
      transactions: [
        {
          id: "tx-1",
          recipient_id: "u2",
          recipient_email: "recipient@example.com",
          amount: 25000,
          created_at: "2026-01-15T10:00:00Z",
        },
      ],
    } as ReturnType<typeof useTransactionHistory>);

    render(<TransactionTable onSelectTransaction={vi.fn()} />);

    expect(screen.getByText("Top Up")).toBeInTheDocument();
  });

  it("updates the date range via setDateRange when the inputs change", () => {
    const setDateRange = vi.fn();
    vi.mocked(useTransactionHistory).mockReturnValue({
      ...baseHookReturn,
      setDateRange,
    } as ReturnType<typeof useTransactionHistory>);
    render(<TransactionTable onSelectTransaction={vi.fn()} />);

    const [startInput] = screen.getAllByDisplayValue("2026-01-01");
    fireEvent.change(startInput, { target: { value: "2026-02-01" } });

    expect(setDateRange).toHaveBeenCalledWith(
      expect.objectContaining({ start_date: "2026-02-01" }),
    );
  });
});
