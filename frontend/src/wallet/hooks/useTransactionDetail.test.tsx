import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useTransactionDetail } from "./useTransactionDetail";
import { getTransactionApi } from "@/wallet/utils/api";

vi.mock("@/wallet/utils/api", () => ({
  getTransactionApi: vi.fn(),
}));

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient();
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("useTransactionDetail", () => {
  beforeEach(() => {
    vi.mocked(getTransactionApi).mockReset();
  });

  it("does not fetch when transactionId is null", () => {
    renderHook(() => useTransactionDetail(null), { wrapper });
    expect(getTransactionApi).not.toHaveBeenCalled();
  });

  it("fetches the transaction once an id is provided", async () => {
    vi.mocked(getTransactionApi).mockResolvedValue({
      message: "ok",
      id: "tx-1",
      recipient_id: "u2",
      recipient_email: "recipient@example.com",
      amount: 25,
      created_at: "2026-01-01T00:00:00Z",
    });

    const { result } = renderHook(() => useTransactionDetail("tx-1"), {
      wrapper,
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(getTransactionApi).toHaveBeenCalledWith("tx-1");
    expect(result.current.data?.recipient_email).toBe("recipient@example.com");
  });
});
