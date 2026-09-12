import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useBalanceQuery } from "./useBalanceQuery";
import { getBalanceApi } from "@/wallet/utils/api";

vi.mock("@/wallet/utils/api", () => ({
  getBalanceApi: vi.fn(),
}));

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient();
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("useBalanceQuery", () => {
  beforeEach(() => {
    vi.mocked(getBalanceApi).mockReset();
  });

  it("returns the balance data once loaded", async () => {
    vi.mocked(getBalanceApi).mockResolvedValue({
      message: "Hello user@example.com, welcome back",
      balance: 1000,
    });

    const { result } = renderHook(() => useBalanceQuery(), { wrapper });

    expect(result.current.isLoading).toBe(true);
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual({
      message: "Hello user@example.com, welcome back",
      balance: 1000,
    });
  });
});
