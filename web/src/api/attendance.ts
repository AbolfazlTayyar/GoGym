import { api } from "./client";
import type { Attendance, AttendanceCount } from "./types";

// Every athlete whose current plan is tracked; the home rows match them by athlete_id.
export const getCurrentAttendance = () => api.get<AttendanceCount[]>("/attendance/current");

export const getAttendance = (planId: string) => api.get<Attendance>(`/plans/${planId}/attendance`);

export const trackPlan = (planId: string) => api.put<Attendance>(`/plans/${planId}/attendance`);

// Deletes the plan's visits too.
export const untrackPlan = (planId: string) =>
  api.delete<Attendance>(`/plans/${planId}/attendance`);

export const addVisit = (planId: string, date: string) =>
  api.post<Attendance>(`/plans/${planId}/attendance/visits`, { date });

export const deleteVisit = (planId: string, date: string) =>
  api.delete<Attendance>(`/plans/${planId}/attendance/visits/${date}`);
