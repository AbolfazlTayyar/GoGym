import type { AthleteType, ExperienceLevel } from "../api/types";

export const athleteTypeLabels: Record<AthleteType, string> = {
  private: "خصوصی",
  public: "عمومی",
};

export const experienceLabels: Record<ExperienceLevel, string> = {
  beginner: "مبتدی",
  intermediate: "متوسط",
  advanced: "حرفه‌ای",
};
