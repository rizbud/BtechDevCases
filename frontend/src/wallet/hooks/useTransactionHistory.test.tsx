import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useTransactionHistory } from "./useTransactionHistory";
import { getTransactionsApi } from "@/wallet/utils/api";

vi.mock("@/wallet/utils/api", () => ({
  getTransactionsApi: vi.fn(),
}));

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient();
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("useTransactionHistory", () => {
  beforeEach(() => {
    vi.mocked(getTransactionsApi).mockReset();
    vi.mocked(getTransactionsApi).mockResolvedValue({
      data: [],
      message: "ok",
      total_records: 0,
      total_pages: 1,
      current_page: 1,
      page_size: 10,
    });
  });

  it("defaults to page 1 with a 30-day range and page size 10", async () => {
    const { result } = renderHook(() => useTransactionHistory(), { wrapper });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.page).toBe(1);
    expect(getTransactionsApi).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, page_size: 10 }),
    );
  });

  it("resets the page to 1 when the date range changes", async () => {
    const { result } = renderHook(() => useTransactionHistory(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    act(() => result.current.setPage(2));
    await waitFor(() => expect(result.current.page).toBe(2));

    act(() =>
      result.current.setDateRange({
        start_date: "2026-02-01",
        end_date: "2026-02-28",
      }),
    );

    expect(result.current.page).toBe(1);
    expect(result.current.dateRange).toEqual({
      start_date: "2026-02-01",
      end_date: "2026-02-28",
    });
  });

  it("exposes transactions and totalPages from the query result", async () => {
    vi.mocked(getTransactionsApi).mockResolvedValue({
      data: [
        {
          id: "tx-1",
          recipient_id: "u2",
          recipient_email: "recipient@example.com",
          amount: 25,
          created_at: "2026-01-01T00:00:00Z",
        },
      ],
      message: "ok",
      total_records: 1,
      total_pages: 2,
      current_page: 1,
      page_size: 10,
    });

    const { result } = renderHook(() => useTransactionHistory(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.transactions).toHaveLength(1);
    expect(result.current.totalPages).toBe(2);
  });
});
