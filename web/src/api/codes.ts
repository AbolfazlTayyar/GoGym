// The error codes the app branches on, mirroring internal/platform/httpx/codes.go, plus two of the client's own.
export const ErrorCode = {
  validationFailed: "validation_failed",
  invalidCredentials: "invalid_credentials",
  unauthorized: "unauthorized",
  notFound: "not_found",
  forbidden: "forbidden",
  conflict: "conflict",
  rateLimited: "rate_limited",
  methodNotAllowed: "method_not_allowed",
  internalError: "internal_error",
  // The request never got an answer: offline, DNS, a dropped connection.
  network: "network_error",
  // An answer that isn't the API's envelope, such as a proxy's 502 page.
  badResponse: "bad_response",
} as const;

// What the coach reads. The server's error.message is English and never shown.
const messages: Record<string, string> = {
  [ErrorCode.validationFailed]: "بعضی از فیلدها درست نیستند.",
  [ErrorCode.invalidCredentials]: "موبایل یا رمز عبور درست نیست.",
  [ErrorCode.unauthorized]: "لطفاً دوباره وارد شوید.",
  [ErrorCode.notFound]: "پیدا نشد؛ شاید حذف شده باشد.",
  [ErrorCode.forbidden]: "این مورد قابل تغییر نیست.",
  [ErrorCode.conflict]: "این تغییر با اطلاعات فعلی جور نیست.",
  [ErrorCode.rateLimited]: "تلاش‌ها زیاد بود؛ چند دقیقه دیگر دوباره امتحان کنید.",
  [ErrorCode.methodNotAllowed]: "این درخواست پشتیبانی نمی‌شود.",
  [ErrorCode.internalError]: "خطایی در سرور رخ داد؛ دوباره امتحان کنید.",
  [ErrorCode.network]: "اتصال برقرار نشد؛ اینترنت را بررسی کنید.",
  [ErrorCode.badResponse]: "سرور در دسترس نیست؛ کمی بعد دوباره امتحان کنید.",
};

const unknownMessage = "خطایی رخ داد؛ دوباره امتحان کنید.";

export function messageForCode(code: string): string {
  return messages[code] ?? unknownMessage;
}
