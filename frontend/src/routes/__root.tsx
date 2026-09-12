import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { useIdleLogout } from "@/auth/hooks/useIdleLogout";

const RootLayout = () => {
  useIdleLogout();
  return (
    <>
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
};

export const Route = createRootRoute({ component: RootLayout });
