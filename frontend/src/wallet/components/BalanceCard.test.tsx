import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BalanceCard } from "./BalanceCard";
import { useBalanceQuery } from "@/wallet/hooks/useBalanceQuery";
import { logout } from "@/auth/utils/api";

vi.mock("@/wallet/hooks/useBalanceQuery", () => ({
  useBalanceQuery: vi.fn(),
}));
vi.mock("@/auth/utils/api", () => ({
  logout: vi.fn(),
}));
const navigate = vi.fn();
vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => navigate,
}));

describe("BalanceCard", () => {
  beforeEach(() => {
    navigate.mockReset();
    vi.mocked(logout).mockReset();
  });

  it("shows a loading state while the balance query is pending", () => {
    vi.mocked(useBalanceQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
    } as ReturnType<typeof useBalanceQuery>);

    const { container } = render(
      <BalanceCard onTopUp={vi.fn()} onTransfer={vi.fn()} />,
    );

    expect(container.querySelectorAll(".skeleton")).toHaveLength(2);
  });

  it("shows the welcome message and formatted balance once loaded", () => {
    vi.mocked(useBalanceQuery).mockReturnValue({
      data: { message: "Hello user@example.com, welcome back", balance: 150000 },
      isLoading: false,
    } as ReturnType<typeof useBalanceQuery>);

    render(<BalanceCard onTopUp={vi.fn()} onTransfer={vi.fn()} />);

    expect(
      screen.getByText("Hello user@example.com, welcome back"),
    ).toBeInTheDocument();
    expect(screen.getByText(/Rp\s?150\.000/)).toBeInTheDocument();
  });

  it("calls onTopUp and onTransfer when their buttons are clicked", async () => {
    vi.mocked(useBalanceQuery).mockReturnValue({
      data: { message: "ok", balance: 0 },
      isLoading: false,
    } as ReturnType<typeof useBalanceQuery>);
    const onTopUp = vi.fn();
    const onTransfer = vi.fn();
    const user = userEvent.setup();
    render(<BalanceCard onTopUp={onTopUp} onTransfer={onTransfer} />);

    await user.click(screen.getByRole("button", { name: "Top Up" }));
    await user.click(screen.getByRole("button", { name: "Transfer" }));

    expect(onTopUp).toHaveBeenCalled();
    expect(onTransfer).toHaveBeenCalled();
  });

  it("logs out and navigates to /login when Logout is clicked", async () => {
    vi.mocked(useBalanceQuery).mockReturnValue({
      data: { message: "ok", balance: 0 },
      isLoading: false,
    } as ReturnType<typeof useBalanceQuery>);
    const user = userEvent.setup();
    render(<BalanceCard onTopUp={vi.fn()} onTransfer={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Logout" }));

    expect(logout).toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ to: "/login" });
  });
});
