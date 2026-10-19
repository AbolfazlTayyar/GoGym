import type { AthleteType } from "../api/types";

// The app's routes, built in one place so links and the router agree.
export const paths = {
  home: "/",
  movements: "/movements",
  // The day view opens on the first day when label is missing.
  dayView: (planId: string, label?: string) =>
    label ? `/plans/${planId}?day=${encodeURIComponent(label)}` : `/plans/${planId}`,
  builder: (planId: string, label?: string) =>
    label ? `/plans/${planId}/edit?day=${encodeURIComponent(label)}` : `/plans/${planId}/edit`,
  profile: (athleteId: string) => `/athletes/${athleteId}`,
  newAthlete: (type: AthleteType) => `/athletes/new?type=${type}`,
  editAthlete: (athleteId: string) => `/athletes/${athleteId}/edit`,
  newPlan: (athleteId: string) => `/athletes/${athleteId}/plans/new`,
};
