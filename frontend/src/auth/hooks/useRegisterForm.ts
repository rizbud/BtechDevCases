import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AxiosError } from "axios";
import {
  registerSchema,
  type RegisterFormValues,
} from "@/auth/utils/schema";
import { useRegisterMutation } from "@/auth/hooks/useRegisterMutation";

export function useRegisterForm() {
  const mutation = useRegisterMutation();
  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync(values);
    } catch (err) {
      if (err instanceof AxiosError) {
        const data = err?.response?.data;
        if (data?.error && typeof data.error === "object") {
          for (const [field, message] of Object.entries(data.error)) {
            form.setError(field as keyof RegisterFormValues, {
              message: message as string,
            });
          }
          return;
        }
        form.setError("root", {
          message: data?.message ?? "Registration failed",
        });
        return;
      }
      form.setError("root", { message: "An unexpected error occurred" });
    }
  });

  return {
    register: form.register,
    errors: form.formState.errors,
    isSubmitting: mutation.isPending,
    isSuccess: mutation.isSuccess,
    onSubmit,
  };
}
