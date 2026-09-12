import { useQuery } from "@tanstack/react-query";
import { getBalanceApi } from "@/wallet/utils/api";

export const useBalanceQuery = () =>
  useQuery({
    queryKey: ["wallet", "balance"],
    queryFn: getBalanceApi,
  });
