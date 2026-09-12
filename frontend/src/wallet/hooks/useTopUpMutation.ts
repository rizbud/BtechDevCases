import { useMutation, useQueryClient } from "@tanstack/react-query";
import { topUpApi } from "@/wallet/utils/api";

export const useTopUpMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ amount, notes }: { amount: number; notes?: string }) =>
      topUpApi(amount, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
};
