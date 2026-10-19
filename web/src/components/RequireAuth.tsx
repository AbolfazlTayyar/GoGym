import { Outlet } from "react-router";
import { useCoach } from "../api/auth";
import { ApiError } from "../api/client";
import { ErrorCode } from "../api/codes";
import { Splash } from "../pages/LoginPage";
import { BigButton } from "./BigButton";
import { EmptyState } from "./EmptyState";
import s from "./RequireAuth.module.css";
import { Screen } from "./Screen";

// Every page but sign-in sits behind this. A 401 is handled by the API client, which sends the
// coach to /login and remembers this page.
export function RequireAuth() {
  const coach = useCoach();

  if (coach.isSuccess) return <Outlet />;
  if (coach.isPending) return <Splash />;
  if (coach.error instanceof ApiError && coach.error.code === ErrorCode.unauthorized) {
    return <Splash />;
  }
  return (
    <Screen>
      <div className={s.error}>
        <EmptyState
          title={coach.error.message}
          action={<BigButton onClick={() => coach.refetch()}>دوباره</BigButton>}
        />
      </div>
    </Screen>
  );
}
