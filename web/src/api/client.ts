import { ErrorCode, messageForCode } from "./codes";
import { translateFieldMessage } from "./fields";

const apiPrefix = "/api/v1";
export const loginPath = "/auth/login";

type ErrorBody = { code: string; message: string; fields?: Record<string, string> };
type Envelope = { success: boolean; data: unknown; error: ErrorBody | null; meta?: unknown };

// Every failed call throws this. message is Persian (codes.ts); fields hold Persian messages keyed by json name.
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields: Record<string, string>;

  constructor(status: number, code: string, fields: Record<string, string> = {}) {
    super(messageForCode(code));
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

let unauthorizedHandler: () => void = () => {};

// The app registers what a 401 does (clear the cache, go to sign-in); the client only reports it.
export function onUnauthorized(handler: () => void) {
  unauthorizedHandler = handler;
}

function isEnvelope(body: unknown): body is Envelope {
  return typeof body === "object" && body !== null && "success" in body && "error" in body;
}

async function send(method: string, path: string, body?: unknown): Promise<Envelope> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let res: Response;
  try {
    res = await fetch(apiPrefix + path, {
      method,
      headers,
      credentials: "same-origin",
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, ErrorCode.network);
  }

  let parsed: unknown = null;
  try {
    parsed = await res.json();
  } catch {
    // Not JSON at all; handled as a bad response below.
  }
  if (!isEnvelope(parsed)) throw new ApiError(res.status, ErrorCode.badResponse);

  if (!parsed.success) {
    const error = parsed.error ?? { code: ErrorCode.badResponse, message: "" };
    if (res.status === 401 && path !== loginPath) unauthorizedHandler();
    const fields = Object.fromEntries(
      Object.entries(error.fields ?? {}).map(([name, msg]) => [name, translateFieldMessage(msg)]),
    );
    throw new ApiError(res.status, error.code, fields);
  }
  return parsed;
}

async function data<T>(method: string, path: string, body?: unknown): Promise<T> {
  return (await send(method, path, body)).data as T;
}

export const api = {
  get: <T>(path: string) => data<T>("GET", path),
  // For list endpoints whose meta (e.g. pagination) the caller needs.
  getWithMeta: async <T, M>(path: string) => {
    const env = await send("GET", path);
    return { data: env.data as T, meta: env.meta as M };
  },
  post: <T>(path: string, body?: unknown) => data<T>("POST", path, body),
  put: <T>(path: string, body?: unknown) => data<T>("PUT", path, body),
  delete: <T = null>(path: string) => data<T>("DELETE", path),
};

// A form shows these under its fields, by json name; anything that isn't an ApiError has none.
export function fieldErrors(err: unknown): Record<string, string> {
  return err instanceof ApiError ? err.fields : {};
}
