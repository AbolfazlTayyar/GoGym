import { describe, expect, it } from "vitest";
import { translateFieldMessage } from "./fields";

describe("translateFieldMessage", () => {
  it.each([
    ["is required", "لازم است."],
    ["must be at most 100 characters", "حداکثر ۱۰۰ نویسه."],
    ["must be between 1 and 50", "باید بین ۱ و ۵۰ باشد."],
    ["must be between 20 and 300 kg", "باید بین ۲۰ و ۳۰۰ باشد."],
    ["must be between 2000-01-01 and 2026-10-20", "این تاریخ مجاز نیست."],
    [
      "at least one of weight, chest, waist, arm, thigh or hip is required",
      "دست‌کم یکی را پر کنید.",
    ],
    ["at least one movement is required", "دست‌کم یک حرکت لازم است."],
    ["must have at most 10 items", "حداکثر ۱۰ مورد."],
    ["something new", "مقدار درست نیست."],
  ])("%s", (english, persian) => {
    expect(translateFieldMessage(english)).toBe(persian);
  });
});
