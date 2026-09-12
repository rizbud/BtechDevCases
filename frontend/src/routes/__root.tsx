import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { useIdleLogout } from "@/auth/hooks/useIdleLogout";
import { OfflineBanner } from "@/components/OfflineBanner";
import { ToastContainer } from "@/components/ToastContainer";

const RootLayout = () => {
  useIdleLogout();
  return (
    <>
      <OfflineBanner />
      <ToastContainer />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
};

export const Route = createRootRoute({ component: RootLayout });
