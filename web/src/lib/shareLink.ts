import { formatJalaliShort, todayISO } from "./jalali";
import { paths } from "./paths";

export const msgCopyFailed = "کپی نشد.";

// The link the athlete opens: this app's own origin, so dev on the LAN and production both work.
export function sharedPlanURL(token: string): string {
  return location.origin + paths.sharedPlan(token);
}

// expires_at is an instant; the coach reads the day it falls on, on the phone's clock.
export function linkCopiedMessage(expiresAt: string): string {
  return `لینک کپی شد — تا ${formatJalaliShort(todayISO(new Date(expiresAt)))} فعال است`;
}
