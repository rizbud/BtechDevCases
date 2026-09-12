import { describe, expect, it, vi, beforeEach } from "vitest";
import api from "@/utils/api";
import {
  getBalanceApi,
  getTransactionApi,
  getTransactionsApi,
  topUpApi,
  transferApi,
} from "./api";

vi.mock("@/utils/api", () => ({
  default: { get: vi.fn(), post: vi.fn() },
}));

beforeEach(() => {
  vi.mocked(api.get).mockReset();
  vi.mocked(api.post).mockReset();
});

describe("getBalanceApi", () => {
  it("fetches the wallet balance", async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { message: "ok", balance: 100 } });
    const result = await getBalanceApi();
    expect(api.get).toHaveBeenCalledWith("/wallet/balance");
    expect(result).toEqual({ message: "ok", balance: 100 });
  });
});

describe("getTransactionsApi", () => {
  it("passes filters as query params", async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { data: [], message: "ok" } });
    const filters = {
      start_date: "2026-01-01",
      end_date: "2026-01-31",
      page: 1,
      page_size: 10,
    };

    await getTransactionsApi(filters);

    expect(api.get).toHaveBeenCalledWith("/wallet/transactions", {
      params: filters,
    });
  });
});

describe("getTransactionApi", () => {
  it("fetches a single transaction by id", async () => {
    vi.mocked(api.get).mockResolvedValue({ data: { message: "ok", id: "tx-1" } });
    await getTransactionApi("tx-1");
    expect(api.get).toHaveBeenCalledWith("/wallet/transaction/tx-1");
  });
});

describe("topUpApi", () => {
  it("posts the amount and optional notes", async () => {
    vi.mocked(api.post).mockResolvedValue({ data: { message: "ok" } });
    await topUpApi(50, "salary");
    expect(api.post).toHaveBeenCalledWith("/wallet/topup", {
      amount: 50,
      notes: "salary",
    });
  });
});

describe("transferApi", () => {
  it("posts the recipient email, amount, and optional notes", async () => {
    vi.mocked(api.post).mockResolvedValue({ data: { message: "ok" } });
    await transferApi("recipient@example.com", 25, "lunch");
    expect(api.post).toHaveBeenCalledWith("/wallet/transfer", {
      to_user_email: "recipient@example.com",
      amount: 25,
      notes: "lunch",
    });
  });
});
