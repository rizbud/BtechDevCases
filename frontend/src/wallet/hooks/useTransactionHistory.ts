import { useState } from "react";
import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { getTransactionsApi } from "@/wallet/utils/api";
import type { TransactionFilters } from "@/wallet/utils/types";

const toDateInput = (date: Date) => date.toISOString().slice(0, 10);

const defaultFilters = (): Omit<TransactionFilters, "page" | "page_size"> => {
  const end = new Date();
  const start = new Date();
  start.setDate(start.getDate() - 30);
  return { start_date: toDateInput(start), end_date: toDateInput(end) };
};

export function useTransactionHistory() {
  const [dateRange, setDateRange] = useState(defaultFilters);
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const filters: TransactionFilters = { ...dateRange, page, page_size: pageSize };

  const query = useQuery({
    queryKey: ["wallet", "transactions", filters],
    queryFn: () => getTransactionsApi(filters),
    placeholderData: keepPreviousData,
  });

  const setDateRangeAndResetPage = (next: {
    start_date: string;
    end_date: string;
  }) => {
    setDateRange(next);
    setPage(1);
  };

  return {
    transactions: query.data?.data ?? [],
    totalPages: query.data?.total_pages ?? 1,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    error: query.error,
    dateRange,
    setDateRange: setDateRangeAndResetPage,
    page,
    setPage,
  };
}
