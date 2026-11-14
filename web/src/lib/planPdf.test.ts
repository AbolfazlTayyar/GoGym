import { describe, expect, it } from "vitest";
import { planPdfFileName } from "./planPdf";

describe("planPdfFileName", () => {
  it("names the athlete and the plan, and ends with the brand", () => {
    expect(planPdfFileName("سارا", "محمدی", "برنامه حجم")).toBe(
      "سارا محمدی - برنامه حجم - gogymapp.pdf",
    );
  });

  it("replaces characters a file name can't hold", () => {
    expect(planPdfFileName("سارا", "محمدی", "هفته ۱/۲")).toBe(
      "سارا محمدی - هفته ۱-۲ - gogymapp.pdf",
    );
  });
});
