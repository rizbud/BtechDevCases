import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AxiosError } from "axios";
import {
  transferSchema,
  type TransferFormInput,
  type TransferFormValues,
} from "@/wallet/utils/schema";
import { useTransferMutation } from "@/wallet/hooks/useTransferMutation";

export function useTransferForm(onSuccess: () => void) {
  const mutation = useTransferMutation();
  const form = useForm<TransferFormInput, unknown, TransferFormValues>({
    resolver: zodResolver(transferSchema),
  });

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      await mutation.mutateAsync({
        toUserEmail: values.to_user_email,
        amount: values.amount,
        notes: values.notes,
      });
      form.reset();
      onSuccess();
    } catch (err) {
      if (err instanceof AxiosError) {
        const data = err?.response?.data;
        if (data?.error && typeof data.error === "object") {
          for (const [field, message] of Object.entries(data.error)) {
            form.setError(field as keyof TransferFormInput, {
              message: message as string,
            });
          }
          return;
        }
        form.setError("root", { message: data?.message ?? "Transfer failed" });
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
