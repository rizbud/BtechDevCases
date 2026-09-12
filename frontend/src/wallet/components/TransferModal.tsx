import { Modal } from "@/components/Modal";
import { useTransferForm } from "@/wallet/hooks/useTransferForm";

interface TransferModalProps {
  open: boolean;
  onClose: () => void;
}

export function TransferModal({ open, onClose }: TransferModalProps) {
  const { register, errors, isSubmitting, onSubmit } = useTransferForm(onClose);

  return (
    <Modal open={open} onClose={onClose} title="Transfer">
      <form onSubmit={onSubmit} className="flex flex-col gap-3">
        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Recipient email</label>
          <input
            type="email"
            placeholder="recipient@example.com"
            className="input input-bordered w-full"
            {...register("to_user_email")}
          />
          {errors.to_user_email && (
            <p className="text-error text-sm">{errors.to_user_email.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Amount</label>
          <input
            type="number"
            step="0.01"
            placeholder="0.00"
            className="input input-bordered w-full"
            {...register("amount")}
          />
          {errors.amount && (
            <p className="text-error text-sm">{errors.amount.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Notes (optional)</label>
          <input
            type="text"
            placeholder="Add a note"
            className="input input-bordered w-full"
            {...register("notes")}
          />
        </div>

        {errors.root && (
          <p className="text-error text-sm">{errors.root.message}</p>
        )}

        <button type="submit" className="btn btn-primary mt-2" disabled={isSubmitting}>
          {isSubmitting ? "Processing..." : "Transfer"}
        </button>
      </form>
    </Modal>
  );
}
