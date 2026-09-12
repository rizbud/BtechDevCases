import {
  createRootRoute,
  Link,
  Outlet,
  useNavigate,
  useRouterState,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { logout } from "@/auth/utils/api";

const AUTH_PATHS = ["/login", "/register"];

const RootLayout = () => {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isLoggedIn = !!localStorage.getItem("token");
  const hideHeader = AUTH_PATHS.includes(pathname);

  const handleLogout = () => {
    logout();
    navigate({ to: "/login" });
  };

  return (
    <>
      {!hideHeader && (
        <>
          <div className="p-2 flex gap-2 items-center">
            <Link to="/" className="[&.active]:font-bold">
              Home
            </Link>
            {isLoggedIn ? (
              <button onClick={handleLogout} className="btn btn-sm ml-auto">
                Logout
              </button>
            ) : (
              <>
                <Link to="/login" className="[&.active]:font-bold ml-auto">
                  Login
                </Link>
                <Link to="/register" className="[&.active]:font-bold">
                  Register
                </Link>
              </>
            )}
          </div>
          <hr />
        </>
      )}
      <Outlet />
      <TanStackRouterDevtools />
    </>
  );
};

export const Route = createRootRoute({ component: RootLayout });
