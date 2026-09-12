import { describe, expect, it } from "vitest";
import { formatCurrency, formatDateTime } from "./format";

describe("formatCurrency", () => {
  it("formats a number as IDR with no decimal places", () => {
    expect(formatCurrency(150000)).toMatch(/^Rp\s?150\.000$/);
  });

  it("formats zero", () => {
    expect(formatCurrency(0)).toMatch(/^Rp\s?0$/);
  });
});

describe("formatDateTime", () => {
  it("formats an ISO date string into a locale string", () => {
    const result = formatDateTime("2026-01-15T10:30:00Z");
    expect(result).toBe(new Date("2026-01-15T10:30:00Z").toLocaleString());
  });
});
