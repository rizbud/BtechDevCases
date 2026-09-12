import { useMutation, useQueryClient } from "@tanstack/react-query";
import { transferApi } from "@/wallet/utils/api";

export const useTransferMutation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      toUserEmail,
      amount,
      notes,
    }: {
      toUserEmail: string;
      amount: number;
      notes?: string;
    }) => transferApi(toUserEmail, amount, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["wallet"] });
    },
  });
};
