import { useNavigate } from "@tanstack/react-router";
import { logout } from "@/auth/utils/api";
import { useBalanceQuery } from "@/wallet/hooks/useBalanceQuery";
import { formatCurrency } from "@/wallet/utils/format";

interface BalanceCardProps {
  onTopUp: () => void;
  onTransfer: () => void;
}

export function BalanceCard({ onTopUp, onTransfer }: BalanceCardProps) {
  const navigate = useNavigate();
  const { data, isLoading } = useBalanceQuery();

  const handleLogout = () => {
    logout();
    navigate({ to: "/login" });
  };

  return (
    <div className="card bg-base-100 shadow-xl p-6 gap-3">
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm text-base-content/60">
          {isLoading ? "Loading..." : data?.message}
        </p>
        <button className="btn btn-sm btn-ghost" onClick={handleLogout}>
          Logout
        </button>
      </div>
      <p className="text-3xl font-bold">
        {isLoading ? "Loading..." : formatCurrency(data?.balance ?? 0)}
      </p>
      <div className="flex gap-2 mt-2">
        <button className="btn btn-primary" onClick={onTopUp}>
          Top Up
        </button>
        <button className="btn btn-outline" onClick={onTransfer}>
          Transfer
        </button>
      </div>
    </div>
  );
}
