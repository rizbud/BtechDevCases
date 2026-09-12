import { z } from "zod";

export const topUpSchema = z.object({
  amount: z.coerce.number().positive("Amount must be greater than zero"),
  notes: z.string().optional(),
});

export type TopUpFormInput = z.input<typeof topUpSchema>;
export type TopUpFormValues = z.output<typeof topUpSchema>;

export const transferSchema = z.object({
  recipient: z
    .string()
    .min(1, "Recipient email is required")
    .email("Invalid email format"),
  amount: z.coerce.number().positive("Amount must be greater than zero"),
  notes: z.string().optional(),
});

export type TransferFormInput = z.input<typeof transferSchema>;
export type TransferFormValues = z.output<typeof transferSchema>;
