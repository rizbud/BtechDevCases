import { useState } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { BalanceCard } from "@/wallet/components/BalanceCard";
import { TopUpModal } from "@/wallet/components/TopUpModal";
import { TransferModal } from "@/wallet/components/TransferModal";
import { TransactionTable } from "@/wallet/components/TransactionTable";
import { TransactionDetailModal } from "@/wallet/components/TransactionDetailModal";

export const Route = createFileRoute("/")({
  component: Wallet,
});

function Wallet() {
  const [isTopUpOpen, setTopUpOpen] = useState(false);
  const [isTransferOpen, setTransferOpen] = useState(false);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(
    null,
  );

  return (
    <div className="p-4 max-w-3xl mx-auto flex flex-col gap-4">
      <BalanceCard
        onTopUp={() => setTopUpOpen(true)}
        onTransfer={() => setTransferOpen(true)}
      />

      <TransactionTable onSelectTransaction={setSelectedTransactionId} />

      <TopUpModal open={isTopUpOpen} onClose={() => setTopUpOpen(false)} />
      <TransferModal open={isTransferOpen} onClose={() => setTransferOpen(false)} />
      <TransactionDetailModal
        transactionId={selectedTransactionId}
        onClose={() => setSelectedTransactionId(null)}
      />
    </div>
  );
}
