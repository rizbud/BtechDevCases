export type Toast = { id: number; message: string; type: "warning" | "error" };

let toasts: Toast[] = [];
let nextId = 0;
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((listener) => listener());
}

export function showToast(message: string, type: Toast["type"] = "warning") {
  const id = nextId++;
  toasts = [...toasts, { id, message, type }];
  emit();
  setTimeout(() => {
    toasts = toasts.filter((t) => t.id !== id);
    emit();
  }, 4000);
}

export function subscribeToasts(callback: () => void) {
  listeners.add(callback);
  return () => listeners.delete(callback);
}

export function getToasts() {
  return toasts;
}
