import { Pagination } from "@/components/Pagination";
import { useTransactionHistory } from "@/wallet/hooks/useTransactionHistory";
import { formatCurrency, formatDateTime } from "@/wallet/utils/format";

interface TransactionTableProps {
  onSelectTransaction: (id: string) => void;
}

export function TransactionTable({ onSelectTransaction }: TransactionTableProps) {
  const {
    transactions,
    totalPages,
    isLoading,
    dateRange,
    setDateRange,
    page,
    setPage,
  } = useTransactionHistory();

  return (
    <div className="card bg-base-100 shadow-xl p-6">
      <div className="flex items-center justify-between mb-4 gap-2 flex-wrap">
        <h3 className="text-lg font-bold">Transaction History</h3>
        <div className="flex gap-2 items-center">
          <input
            type="date"
            className="input input-bordered input-sm"
            value={dateRange.start_date}
            max={dateRange.end_date}
            onChange={(e) =>
              setDateRange({ ...dateRange, start_date: e.target.value })
            }
          />
          <span>to</span>
          <input
            type="date"
            className="input input-bordered input-sm"
            value={dateRange.end_date}
            min={dateRange.start_date}
            onChange={(e) =>
              setDateRange({ ...dateRange, end_date: e.target.value })
            }
          />
        </div>
      </div>

      {isLoading ? (
        <p>Loading...</p>
      ) : transactions.length === 0 ? (
        <p className="text-base-content/60">No transactions in this range.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Date</th>
                <th>From</th>
                <th>To</th>
                <th className="text-right">Amount</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((transaction) => (
                <tr
                  key={transaction.id}
                  className="hover cursor-pointer"
                  onClick={() => onSelectTransaction(transaction.id)}
                >
                  <td>{formatDateTime(transaction.created_at)}</td>
                  <td>{transaction.sender_email ?? "Top Up"}</td>
                  <td>{transaction.recipient_email}</td>
                  <td className="text-right">{formatCurrency(transaction.amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  );
}
