import { Modal } from "@/components/Modal";
import { useTopUpForm } from "@/wallet/hooks/useTopUpForm";

interface TopUpModalProps {
  open: boolean;
  onClose: () => void;
}

export function TopUpModal({ open, onClose }: TopUpModalProps) {
  const { register, errors, isSubmitting, onSubmit } = useTopUpForm(onClose);

  return (
    <Modal open={open} onClose={onClose} title="Top Up">
      <form onSubmit={onSubmit} className="flex flex-col gap-3">
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
          {isSubmitting ? "Processing..." : "Top Up"}
        </button>
      </form>
    </Modal>
  );
}
