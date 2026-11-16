import { api } from "./client";
import type { MovementCategory } from "./types";

// A category is a name on the coach's movements, so it's addressed by that name. Renaming onto a
// name already in use merges the two.
const categoryPath = (name: string) => `/movement-categories/${encodeURIComponent(name)}`;

export const renameMovementCategory = (name: string, newName: string) =>
  api.put<MovementCategory>(categoryPath(name), { name: newName });

// Leaves the coach's movements in it without a category.
export const deleteMovementCategory = (name: string) => api.delete(categoryPath(name));
