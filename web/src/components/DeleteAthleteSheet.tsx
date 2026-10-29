import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteAthlete } from "../api/athletes";
import { queryKeys } from "../api/keys";
import type { Athlete } from "../api/types";
import { ConfirmSheet } from "./ConfirmSheet";

type DeleteAthleteSheetProps = {
  open: boolean;
  onClose: () => void;
  athlete: Athlete;
  onDeleted: () => void;
};

// Asks before deleting an athlete, then deletes them and refreshes the lists.
export function DeleteAthleteSheet({ open, onClose, athlete, onDeleted }: DeleteAthleteSheetProps) {
  const queryClient = useQueryClient();
  const remove = useMutation({
    mutationFn: () => deleteAthlete(athlete.id),
    onSuccess: async () => {
      // The plan rows on home come from the current plans, which a deleted athlete leaves too.
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.athletes() }),
        queryClient.invalidateQueries({ queryKey: queryKeys.plans() }),
      ]);
      onDeleted();
    },
  });
  return (
    <ConfirmSheet
      open={open}
      onClose={onClose}
      title={`حذف ${athlete.first_name} ${athlete.last_name}؟`}
      text="شاگرد از فهرست شما حذف می‌شود و برنامه‌ها و اندازه‌هایش دیگر دیده نمی‌شوند."
      confirmLabel="حذف"
      onConfirm={() => remove.mutate()}
      pending={remove.isPending}
      error={remove.isError ? remove.error.message : null}
    />
  );
}
