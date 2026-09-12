import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Pagination } from "./Pagination";

describe("Pagination", () => {
  it("renders nothing when there is one page or fewer", () => {
    const { container } = render(
      <Pagination page={1} totalPages={1} onPageChange={vi.fn()} />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("disables the previous button on the first page and the next button on the last page", () => {
    render(<Pagination page={1} totalPages={3} onPageChange={vi.fn()} />);
    expect(screen.getByText("«")).toBeDisabled();
    expect(screen.getByText("»")).not.toBeDisabled();

    render(<Pagination page={3} totalPages={3} onPageChange={vi.fn()} />);
    expect(screen.getAllByText("»")[1]).toBeDisabled();
  });

  it("calls onPageChange with the adjacent page", async () => {
    const onPageChange = vi.fn();
    const user = userEvent.setup();
    render(<Pagination page={2} totalPages={3} onPageChange={onPageChange} />);

    await user.click(screen.getByText("»"));
    expect(onPageChange).toHaveBeenCalledWith(3);

    await user.click(screen.getByText("«"));
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("shows the current page and total pages", () => {
    render(<Pagination page={2} totalPages={5} onPageChange={vi.fn()} />);
    expect(screen.getByText("Page 2 / 5")).toBeInTheDocument();
  });
});
