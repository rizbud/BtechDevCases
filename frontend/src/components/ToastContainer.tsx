import { useSyncExternalStore } from "react";
import { getToasts, subscribeToasts } from "@/utils/toast";

const emptyToasts: ReturnType<typeof getToasts> = [];

export function ToastContainer() {
  const toasts = useSyncExternalStore(
    subscribeToasts,
    getToasts,
    () => emptyToasts,
  );

  if (toasts.length === 0) return null;

  return (
    <div className="toast toast-top toast-center z-50">
      {toasts.map((toast) => (
        <div key={toast.id} className={`alert alert-${toast.type} text-sm`}>
          {toast.message}
        </div>
      ))}
    </div>
  );
}
