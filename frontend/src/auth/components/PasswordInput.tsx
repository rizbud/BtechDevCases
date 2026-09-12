import { useState, type InputHTMLAttributes } from "react";

export function PasswordInput(props: InputHTMLAttributes<HTMLInputElement>) {
  const [visible, setVisible] = useState(false);

  return (
    <div className="relative">
      <input
        {...props}
        type={visible ? "text" : "password"}
        className="input input-bordered w-full pr-12"
      />
      <button
        type="button"
        onClick={() => setVisible((v) => !v)}
        className="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-base-content/60"
        tabIndex={-1}
      >
        {visible ? "Hide" : "Show"}
      </button>
    </div>
  );
}
