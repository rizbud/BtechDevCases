import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { useIdleLogout } from "@/auth/hooks/useIdleLogout";
import { OfflineBanner } from "@/components/OfflineBanner";

const RootLayout = () => {
  useIdleLogout();
  return (
    <>
      <OfflineBanner />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
};

export const Route = createRootRoute({ component: RootLayout });
