import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { logout, useCoach } from "../api/auth";
import { signInPath } from "../api/session";
import { toPersianDigits } from "../lib/digits";
import { BigButton } from "./BigButton";
import { BottomSheet } from "./BottomSheet";
import s from "./SignOutSheet.module.css";

export function SignOutSheet({ open, onClose }: { open: boolean; onClose: () => void }) {
  const coach = useCoach();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  // Replacing the sheet's history entry, so back from sign-in doesn't reopen the app signed out.
  const signOut = useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.clear();
      navigate(signInPath, { replace: true });
    },
  });

  const name = coach.data ? `${coach.data.first_name} ${coach.data.last_name}` : "حساب";

  return (
    <BottomSheet open={open} onClose={onClose} title={name}>
      {coach.data && <p className={s.phone}>{toPersianDigits(coach.data.phone)}</p>}
      <BigButton outline onClick={() => signOut.mutate()} disabled={signOut.isPending}>
        خروج از حساب
      </BigButton>
      {signOut.isError && (
        <p className={s.error} role="alert">
          {signOut.error.message}
        </p>
      )}
    </BottomSheet>
  );
}
