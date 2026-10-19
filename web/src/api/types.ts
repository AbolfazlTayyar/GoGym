// Payload types, written by hand to match the json tags in internal/*/handler.go (docs/adr/0030 says why).
// Nullable fields are `T | null`, never optional: the API always sends the key.

export type Coach = {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
};

export type ListMeta = {
  total: number;
  limit: number;
  offset: number;
};

export type AthleteType = "private" | "public";
export type ExperienceLevel = "beginner" | "intermediate" | "advanced";

export type Athlete = {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
  experience_level: ExperienceLevel | null;
  injuries: string | null;
  goal: string | null;
  height: number | null;
  athlete_type: AthleteType;
  created_at: string;
  updated_at: string;
};

export type Measurement = {
  id: string;
  athlete_id: string;
  date: string;
  weight: number | null;
  chest: number | null;
  waist: number | null;
  arm: number | null;
  thigh: number | null;
  hip: number | null;
  created_at: string;
  updated_at: string;
};

export type Movement = {
  id: string;
  // null for the universal, read-only movements.
  coach_id: string | null;
  name: string;
  category: string | null;
  description: string | null;
  media_url: string | null;
  created_at: string;
  updated_at: string;
};

export type PlanSummary = {
  id: string;
  athlete_id: string;
  title: string;
  note: string | null;
  start_date: string;
  is_current: boolean;
  // In day order; [] for an empty plan.
  day_labels: string[];
  // Numbered lines across all days: a superset counts once.
  block_count: number;
  created_at: string;
  updated_at: string;
};

export type BlockMovement = {
  id: string;
  movement_id: string;
  name: string;
  category: string | null;
  reps: number | null;
  duration_seconds: number | null;
  load: string | null;
  order_in_block: number;
};

export type Block = {
  id: string;
  order_index: number;
  sets: number;
  rest_seconds: number | null;
  notes: string | null;
  movements: BlockMovement[];
};

export type Day = {
  id: string;
  label: string;
  order_index: number;
  blocks: Block[];
};

export type PlanDetail = {
  id: string;
  athlete_id: string;
  title: string;
  note: string | null;
  start_date: string;
  created_at: string;
  updated_at: string;
  // The plan this one was copied from; null when it wasn't a copy or the source was deleted.
  source_plan_id: string | null;
  source_plan_title: string | null;
  days: Day[];
};
