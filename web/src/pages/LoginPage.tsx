import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { Navigate, useSearchParams } from "react-router";
import { login, useCoach } from "../api/auth";
import { ErrorCode } from "../api/codes";
import { ApiError, fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import { nextParam, safeNext } from "../api/session";
import { BigButton } from "../components/BigButton";
import { DumbbellIcon, EyeIcon, EyeOffIcon } from "../components/icons";
import { Screen } from "../components/Screen";
import { TextField } from "../components/TextField";
import { limitPhoneInput, toLatinDigits } from "../lib/digits";
import { focusNextOnEnter } from "../lib/focusNext";
import { msgPhoneEmpty } from "../lib/formMessages";
import s from "./LoginPage.module.css";

// Shown under the form; other failures (a field left empty) show under their field.
const formErrorCodes: string[] = [
  ErrorCode.invalidCredentials,
  ErrorCode.rateLimited,
  ErrorCode.network,
  ErrorCode.badResponse,
  ErrorCode.internalError,
];

export function LoginPage() {
  const coach = useCoach();
  const [params] = useSearchParams();

  // Only the logo until the API says whether the cookie is still good, so a signed-in coach never sees the form.
  if (coach.isPending) return <Splash />;
  if (coach.isSuccess) return <Navigate to={safeNext(params.get(nextParam))} replace />;
  return <LoginForm />;
}

function Brand() {
  return (
    <>
      <div className={s.logo}>
        <DumbbellIcon />
      </div>
      <div className={s.brand}>GoGym</div>
    </>
  );
}

export function Splash() {
  return (
    <Screen>
      <div className={s.login} aria-busy>
        <Brand />
      </div>
    </Screen>
  );
}

function LoginForm() {
  const queryClient = useQueryClient();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [emptyErrors, setEmptyErrors] = useState<Record<string, string>>({});

  // On success the coach query refetches, and LoginPage sends the coach on to where they were going.
  const submit = useMutation({
    mutationFn: () => login(toLatinDigits(phone.trim()), password),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.coach() }),
  });
  const errors = { ...fieldErrors(submit.error), ...emptyErrors };
  const formError =
    submit.error instanceof ApiError && formErrorCodes.includes(submit.error.code)
      ? submit.error.message
      : null;

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    const empty: Record<string, string> = {};
    if (!phone.trim()) empty.phone = msgPhoneEmpty;
    if (!password) empty.password = "رمز عبور را وارد کنید.";
    setEmptyErrors(empty);
    if (Object.keys(empty).length > 0) {
      submit.reset();
      return;
    }
    submit.mutate();
  };

  return (
    <Screen>
      <form className={s.login} onSubmit={onSubmit} onKeyDown={focusNextOnEnter} noValidate>
        <Brand />
        <p className={s.sub}>ورود مربی</p>
        <div className={s.fields}>
          <TextField
            label="موبایل"
            type="tel"
            inputMode="numeric"
            autoComplete="username"
            placeholder="۰۹۱۲ ۱۲۳ ۴۵۶۷"
            ltr
            value={phone}
            onChange={(e) => setPhone(limitPhoneInput(e.target.value))}
            error={errors.phone}
          />
          <TextField
            label="رمز عبور"
            type={showPassword ? "text" : "password"}
            autoComplete="current-password"
            ltr
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            trailing={
              <button
                type="button"
                className={s.eye}
                aria-label={showPassword ? "پنهان کردن رمز" : "نمایش رمز"}
                aria-pressed={showPassword}
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? <EyeOffIcon /> : <EyeIcon />}
              </button>
            }
          />
        </div>
        {formError && (
          <p className={s.formError} role="alert">
            {formError}
          </p>
        )}
        <div className={s.bottom}>
          <BigButton type="submit" disabled={submit.isPending}>
            ورود
          </BigButton>
        </div>
      </form>
    </Screen>
  );
}
