import api from "@/utils/api";
import type {
  PaginationMeta,
  Transaction,
  TransactionFilters,
} from "@/wallet/utils/types";

export const getBalanceApi = async () => {
  const response = await api.get<{ message: string; balance: number }>(
    "/wallet/balance",
  );
  return response.data;
};

export const getTransactionsApi = async (filters: TransactionFilters) => {
  const response = await api.get<
    { data: Transaction[]; message: string } & PaginationMeta
  >("/wallet/transactions", { params: filters });
  return response.data;
};

export const getTransactionApi = async (transactionId: string) => {
  const response = await api.get<{ message: string } & Transaction>(
    `/wallet/transaction/${transactionId}`,
  );
  return response.data;
};

export const topUpApi = async (amount: number, notes?: string) => {
  const response = await api.post<{ message: string } & Transaction>(
    "/wallet/topup",
    { amount, notes },
  );
  return response.data;
};

export const transferApi = async (
  toUserEmail: string,
  amount: number,
  notes?: string,
) => {
  const response = await api.post<{ message: string } & Transaction>(
    "/wallet/transfer",
    { recipient: toUserEmail, amount, notes },
  );
  return response.data;
};
