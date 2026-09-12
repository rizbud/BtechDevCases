import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { AxiosError } from "axios";
import { loginSchema, type LoginFormValues } from "@/auth/utils/schema";
import { useLoginMutation } from "@/auth/hooks/useLoginMutation";

export function useLoginForm() {
  const navigate = useNavigate();
  const mutation = useLoginMutation();
  const form = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) });
  const [isRedirecting, setIsRedirecting] = useState(false);

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync(values);
      setIsRedirecting(true);
      await navigate({ to: "/" });
    } catch (err) {
      setIsRedirecting(false);
      if (err instanceof AxiosError) {
        const data = err?.response?.data;
        if (data?.error && typeof data.error === "object") {
          for (const [field, message] of Object.entries(data.error)) {
            form.setError(field as keyof LoginFormValues, {
              message: message as string,
            });
          }
          return;
        }
        form.setError("root", { message: data?.message ?? "Login failed" });
        return;
      }
      form.setError("root", { message: "An unexpected error occurred" });
    }
  });

  return {
    register: form.register,
    errors: form.formState.errors,
    isSubmitting: mutation.isPending || isRedirecting,
    onSubmit,
  };
}
