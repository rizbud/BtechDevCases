import { useMutation, useQueryClient } from "@tanstack/react-query";
import { transferApi } from "@/wallet/utils/api";

export const useTransferMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ toUserEmail, amount }: { toUserEmail: string; amount: number }) =>
      transferApi(toUserEmail, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
};
