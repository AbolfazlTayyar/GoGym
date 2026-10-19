// One key function per resource, so a save can invalidate everything under a prefix, e.g. queryKeys.athletes().
export const queryKeys = {
  coach: () => ["coach"] as const,
  athletes: (...parts: unknown[]) => ["athletes", ...parts] as const,
  plans: (...parts: unknown[]) => ["plans", ...parts] as const,
  measurements: (...parts: unknown[]) => ["measurements", ...parts] as const,
  movements: (...parts: unknown[]) => ["movements", ...parts] as const,
};
