import { useQuery } from "@tanstack/react-query";
import { getTransactionApi } from "@/wallet/utils/api";

export function useTransactionDetail(transactionId: string | null) {
  return useQuery({
    queryKey: ["wallet", "transaction", transactionId],
    queryFn: () => getTransactionApi(transactionId as string),
    enabled: transactionId !== null,
  });
}
