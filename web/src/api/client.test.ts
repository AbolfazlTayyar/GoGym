import { QueryClient } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, fieldErrors, loginPath, onUnauthorized } from "./client";
import { ErrorCode } from "./codes";
import { shouldRetry } from "./queryClient";
import { goToSignIn } from "./session";

const fetchMock = vi.fn();

function respond(status: number, body: unknown) {
  fetchMock.mockResolvedValueOnce(
    new Response(typeof body === "string" ? body : JSON.stringify(body), { status }),
  );
}

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  fetchMock.mockReset();
  vi.unstubAllGlobals();
  onUnauthorized(() => {});
});

describe("api client", () => {
  it("returns data, and meta when asked for it", async () => {
    const meta = { total: 2, limit: 20, offset: 0 };
    respond(200, { success: true, data: [{ id: "a" }, { id: "b" }], error: null, meta });

    const result = await api.getWithMeta("/athletes");

    expect(result).toEqual({ data: [{ id: "a" }, { id: "b" }], meta });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/athletes",
      expect.objectContaining({ method: "GET", credentials: "same-origin" }),
    );
  });

  it("sends a JSON body and returns only data", async () => {
    respond(201, { success: true, data: { id: "a" }, error: null });

    await expect(api.post("/athletes", { first_name: "سارا" })).resolves.toEqual({ id: "a" });
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(init.body).toBe('{"first_name":"سارا"}');
    expect(init.headers).toMatchObject({ "Content-Type": "application/json" });
  });

  it("throws validation_failed with Persian per-field messages", async () => {
    respond(400, {
      success: false,
      data: null,
      error: {
        code: "validation_failed",
        message: "invalid request",
        fields: {
          phone: "must be an Iranian mobile number, e.g. 09372144430",
          first_name: "is required",
        },
      },
    });

    const err = await api.post("/athletes", {}).catch((e: unknown) => e);

    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 400, code: ErrorCode.validationFailed });
    expect(fieldErrors(err)).toEqual({
      phone: "شماره موبایل درست نیست؛ مثلاً ۰۹۱۲۳۴۵۶۷۸۹.",
      first_name: "لازم است.",
    });
    expect((err as ApiError).message).not.toMatch(/invalid request/);
  });

  it("sends the app to sign-in on a 401, remembering the page", async () => {
    const queryClient = new QueryClient();
    queryClient.setQueryData(["athletes"], [{ id: "a" }]);
    const navigate = vi.fn();
    onUnauthorized(() => goToSignIn(navigate, queryClient, "/movements?q=x"));
    respond(401, {
      success: false,
      data: null,
      error: { code: "unauthorized", message: "unauthorized" },
    });

    await expect(api.get("/movements")).rejects.toMatchObject({ code: ErrorCode.unauthorized });

    expect(navigate).toHaveBeenCalledWith("/login?next=%2Fmovements%3Fq%3Dx", { replace: true });
    expect(queryClient.getQueryData(["athletes"])).toBeUndefined();
  });

  it("leaves a 401 from login to the sign-in form", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    respond(401, {
      success: false,
      data: null,
      error: { code: "invalid_credentials", message: "invalid credentials" },
    });

    await expect(api.post(loginPath, {})).rejects.toMatchObject({
      code: ErrorCode.invalidCredentials,
      message: "موبایل یا رمز عبور درست نیست.",
    });
    expect(handler).not.toHaveBeenCalled();
  });

  it("turns a non-JSON 502 into an ApiError", async () => {
    respond(502, "<html><body>Bad Gateway</body></html>");

    await expect(api.get("/athletes")).rejects.toMatchObject({
      status: 502,
      code: ErrorCode.badResponse,
    });
  });

  it("turns JSON that isn't the envelope into an ApiError", async () => {
    respond(200, { status: "ok" });

    await expect(api.get("/athletes")).rejects.toMatchObject({ code: ErrorCode.badResponse });
  });

  it("turns a dropped connection into a network ApiError", async () => {
    fetchMock.mockRejectedValueOnce(new TypeError("Failed to fetch"));

    await expect(api.get("/athletes")).rejects.toMatchObject({
      status: 0,
      code: ErrorCode.network,
    });
  });
});

describe("shouldRetry", () => {
  it("retries network errors only", () => {
    expect(shouldRetry(0, new ApiError(0, ErrorCode.network))).toBe(true);
    expect(shouldRetry(2, new ApiError(0, ErrorCode.network))).toBe(false);
    expect(shouldRetry(0, new ApiError(404, ErrorCode.notFound))).toBe(false);
    expect(shouldRetry(0, new ApiError(500, ErrorCode.internalError))).toBe(false);
  });
});
