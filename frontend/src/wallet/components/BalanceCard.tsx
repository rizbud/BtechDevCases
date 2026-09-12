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
  const { data, isLoading, isError, refetch } = useBalanceQuery();

  const handleLogout = () => {
    logout();
    navigate({ to: "/login" });
  };

  if (isError) {
    return (
      <div className="card bg-base-100 shadow-xl p-6 gap-3">
        <p className="text-sm text-error">
          Couldn't load your balance. Check your connection and try again.
        </p>
        <button className="btn btn-sm" onClick={() => refetch()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="card bg-base-100 shadow-xl p-6 gap-3">
      <div className="flex items-center justify-between gap-2">
        {isLoading ? (
          <div className="skeleton h-4 w-40" />
        ) : (
          <p className="text-sm text-base-content/60">{data?.message}</p>
        )}
        <button className="btn btn-sm btn-ghost" onClick={handleLogout}>
          Logout
        </button>
      </div>
      {isLoading ? (
        <div className="skeleton h-9 w-32" />
      ) : (
        <p className="text-3xl font-bold">
          {formatCurrency(data?.balance ?? 0)}
        </p>
      )}
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
