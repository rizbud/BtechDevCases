import { useMutation } from "@tanstack/react-query";
import { loginApi } from "@/auth/utils/api";

export const useLoginMutation = () =>
  useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginApi(email, password),
  });
