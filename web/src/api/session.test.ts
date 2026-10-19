import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { goToSignIn, safeNext } from "./session";

describe("goToSignIn", () => {
  it("does nothing on the sign-in page, where a 401 is the expected answer", () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(["coach"], null);
    const navigate = vi.fn();

    goToSignIn(navigate, queryClient, "/login?next=%2F");

    expect(navigate).not.toHaveBeenCalled();
    expect(queryClient.getQueryCache().getAll()).toHaveLength(1);
  });
});

describe("safeNext", () => {
  it.each([
    ["/movements", "/movements"],
    ["/plans/1?day=B", "/plans/1?day=B"],
    [null, "/"],
    ["", "/"],
    ["https://evil.example", "/"],
    ["//evil.example", "/"],
  ])("%j → %j", (next, expected) => {
    expect(safeNext(next)).toBe(expected);
  });
});
