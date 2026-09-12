import { useMutation } from "@tanstack/react-query";
import { registerApi } from "@/auth/utils/api";

export const useRegisterMutation = () =>
  useMutation({
    mutationFn: ({
      email,
      password,
      confirmPassword,
    }: {
      email: string;
      password: string;
      confirmPassword: string;
    }) => registerApi(email, password, confirmPassword),
  });
