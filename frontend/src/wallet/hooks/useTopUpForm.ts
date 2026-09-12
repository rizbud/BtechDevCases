import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AxiosError } from "axios";
import {
  topUpSchema,
  type TopUpFormInput,
  type TopUpFormValues,
} from "@/wallet/utils/schema";
import { useTopUpMutation } from "@/wallet/hooks/useTopUpMutation";

export function useTopUpForm(onSuccess: () => void) {
  const mutation = useTopUpMutation();
  const form = useForm<TopUpFormInput, unknown, TopUpFormValues>({
    resolver: zodResolver(topUpSchema),
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync(values.amount);
      form.reset();
      onSuccess();
    } catch (err) {
      if (err instanceof AxiosError) {
        const data = err?.response?.data;
        if (data?.error && typeof data.error === "object") {
          for (const [field, message] of Object.entries(data.error)) {
            form.setError(field as keyof TopUpFormInput, {
              message: message as string,
            });
          }
          return;
        }
        form.setError("root", { message: data?.message ?? "Top up failed" });
        return;
      }
      form.setError("root", { message: "An unexpected error occurred" });
    }
  });

  return {
    register: form.register,
    errors: form.formState.errors,
    isSubmitting: mutation.isPending,
    onSubmit,
  };
}
