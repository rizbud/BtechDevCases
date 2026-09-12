import { describe, expect, it } from "vitest";
import { topUpSchema, transferSchema } from "./schema";

describe("topUpSchema", () => {
  it("accepts a positive amount with no notes", () => {
    const result = topUpSchema.safeParse({ amount: "100" });
    expect(result.success).toBe(true);
  });

  it("rejects a zero or negative amount", () => {
    expect(topUpSchema.safeParse({ amount: "0" }).success).toBe(false);
    expect(topUpSchema.safeParse({ amount: "-5" }).success).toBe(false);
  });
});

describe("transferSchema", () => {
  it("accepts a valid recipient, amount, and notes", () => {
    const result = transferSchema.safeParse({
      recipient: "recipient@example.com",
      amount: "50",
      notes: "for lunch",
    });
    expect(result.success).toBe(true);
  });

  it("accepts a missing notes field", () => {
    const result = transferSchema.safeParse({
      recipient: "recipient@example.com",
      amount: "50",
    });
    expect(result.success).toBe(true);
  });

  it("rejects an invalid recipient email", () => {
    const result = transferSchema.safeParse({
      recipient: "not-an-email",
      amount: "50",
    });
    expect(result.success).toBe(false);
  });

  it("rejects a non-positive amount", () => {
    const result = transferSchema.safeParse({
      recipient: "recipient@example.com",
      amount: "0",
    });
    expect(result.success).toBe(false);
  });
});
