import { describe, expect, it } from "vitest";
import { linkCopiedMessage } from "./shareLink";

describe("linkCopiedMessage", () => {
  it("names the Jalali day the link expires on", () => {
    // Midday UTC, so the phone's time zone can't move it to another day.
    expect(linkCopiedMessage("2026-12-14T12:00:00Z")).toBe("لینک کپی شد — تا ۲۳ آذر فعال است");
  });
});
