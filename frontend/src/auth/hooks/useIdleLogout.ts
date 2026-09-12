import { useEffect, useRef } from "react";
import { useNavigate } from "@tanstack/react-router";
import { logout } from "@/auth/utils/api";

export const IDLE_TIMEOUT_MS = 15 * 60 * 1000;

const ACTIVITY_EVENTS = [
  "mousemove",
  "mousedown",
  "keydown",
  "touchstart",
  "scroll",
] as const;

export function useIdleLogout() {
  const navigate = useNavigate();
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined,
  );

  useEffect(() => {
    const handleIdle = () => {
      if (!localStorage.getItem("token")) return;
      logout();
      navigate({ to: "/login" });
    };

    const resetTimer = () => {
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(handleIdle, IDLE_TIMEOUT_MS);
    };

    resetTimer();
    for (const event of ACTIVITY_EVENTS) {
      window.addEventListener(event, resetTimer);
    }

    return () => {
      clearTimeout(timerRef.current);
      for (const event of ACTIVITY_EVENTS) {
        window.removeEventListener(event, resetTimer);
      }
    };
  }, [navigate]);
}
