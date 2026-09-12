import { Modal } from "@/components/Modal";
import { useTransactionDetail } from "@/wallet/hooks/useTransactionDetail";
import { formatCurrency, formatDateTime } from "@/wallet/utils/format";

interface TransactionDetailModalProps {
  transactionId: string | null;
  onClose: () => void;
}

export function TransactionDetailModal({
  transactionId,
  onClose,
}: TransactionDetailModalProps) {
  const { data, isLoading } = useTransactionDetail(transactionId);

  return (
    <Modal
      open={transactionId !== null}
      onClose={onClose}
      title="Transaction Detail"
    >
      {isLoading || !data ? (
        <p>Loading...</p>
      ) : (
        <div className="flex flex-col gap-2 text-sm">
          <div className="flex justify-between">
            <span className="text-base-content/60">ID</span>
            <span className="font-mono">{data.id}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-base-content/60">From</span>
            <span>{data.sender_email ?? "Top Up"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-base-content/60">To</span>
            <span>{data.recipient_email}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-base-content/60">Amount</span>
            <span className="font-bold">{formatCurrency(data.amount)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-base-content/60">Date</span>
            <span>{formatDateTime(data.created_at)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-base-content/60">Notes</span>
            <span>{data.notes || "-"}</span>
          </div>
        </div>
      )}
    </Modal>
  );
}
