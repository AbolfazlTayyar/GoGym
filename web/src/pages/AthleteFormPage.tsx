import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { createAthlete, getAthlete, updateAthlete, type AthleteInput } from "../api/athletes";
import { fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import type { Athlete, AthleteType, ExperienceLevel } from "../api/types";
import { BigButton } from "../components/BigButton";
import { ConfirmSheet } from "../components/ConfirmSheet";
import { DeleteAthleteSheet } from "../components/DeleteAthleteSheet";
import { EmptyState } from "../components/EmptyState";
import f from "../components/Field.module.css";
import { Header, HeaderTitle, PillButton, RoundButton } from "../components/Header";
import { TrashIcon, XIcon } from "../components/icons";
import { Screen, ScreenContent } from "../components/Screen";
import { Segmented } from "../components/Segmented";
import { TextAreaField, TextField } from "../components/TextField";
import { useSheet } from "../components/useSheet";
import { limitPhoneInput, parseNumber, toPersianDigits, toPhone } from "../lib/digits";
import { msgNotANumber, msgPhoneEmpty } from "../lib/formMessages";
import { focusNextOnEnter } from "../lib/focusNext";
import { athleteTypeLabels, experienceLabels } from "../lib/labels";
import { paths } from "../lib/paths";
import { useGoBack } from "../lib/useGoBack";
import s from "./AthleteFormPage.module.css";

// What the coach types, as typed; toInput turns it into the API's shape.
type FormState = {
  firstName: string;
  lastName: string;
  phone: string;
  type: AthleteType;
  level: ExperienceLevel | null;
  goal: string;
  height: string;
  injuries: string;
};

function emptyForm(type: AthleteType): FormState {
  return {
    firstName: "",
    lastName: "",
    phone: "",
    type,
    level: null,
    goal: "",
    height: "",
    injuries: "",
  };
}

function formFrom(a: Athlete): FormState {
  return {
    firstName: a.first_name,
    lastName: a.last_name,
    phone: toPersianDigits(a.phone),
    type: a.athlete_type,
    level: a.experience_level,
    goal: a.goal ?? "",
    height: a.height === null ? "" : toPersianDigits(a.height),
    injuries: a.injuries ?? "",
  };
}

function toInput(form: FormState): AthleteInput {
  return {
    first_name: form.firstName.trim(),
    last_name: form.lastName.trim(),
    phone: toPhone(form.phone),
    athlete_type: form.type,
    experience_level: form.level,
    goal: form.goal.trim() || null,
    height: parseNumber(form.height),
    injuries: form.injuries.trim() || null,
  };
}

export function NewAthletePage() {
  const [params] = useSearchParams();
  const type: AthleteType = params.get("type") === "public" ? "public" : "private";
  return <AthleteForm initial={emptyForm(type)} />;
}

export function EditAthletePage() {
  const { id = "" } = useParams();
  const athlete = useQuery({
    queryKey: queryKeys.athletes("detail", id),
    queryFn: () => getAthlete(id),
  });

  if (athlete.isPending) return <Screen>{null}</Screen>;
  if (athlete.isError) {
    return (
      <Screen>
        <EmptyState
          title={athlete.error.message}
          action={<BigButton to={paths.home}>بازگشت به شاگردها</BigButton>}
        />
      </Screen>
    );
  }
  // Keyed by id so the form starts from this athlete's saved values.
  return <AthleteForm key={id} athlete={athlete.data} initial={formFrom(athlete.data)} />;
}

function AthleteForm({ athlete, initial }: { athlete?: Athlete; initial: FormState }) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [form, setForm] = useState(initial);
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [localErrors, setLocalErrors] = useState<Record<string, string>>({});
  const discard = useSheet("discard");
  const remove = useSheet("delete");
  const goBack = useGoBack(athlete ? paths.profile(athlete.id) : paths.home);
  const set = <K extends keyof FormState>(key: K, value: FormState[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const dirty = JSON.stringify(form) !== JSON.stringify(initial);

  const save = useMutation({
    mutationFn: (input: AthleteInput) =>
      athlete ? updateAthlete(athlete.id, input) : createAthlete(input),
    onSuccess: async (saved) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.athletes() });
      // Replacing the form, so back from the profile doesn't reopen it.
      navigate(paths.profile(saved.id), { replace: true });
    },
  });

  const errors = { ...fieldErrors(save.error), ...localErrors };

  const onSubmit = (e?: FormEvent) => {
    e?.preventDefault();
    const local: Record<string, string> = {};
    if (!form.firstName.trim()) local.first_name = "نام را وارد کنید.";
    if (!form.lastName.trim()) local.last_name = "نام خانوادگی را وارد کنید.";
    if (!form.phone.trim()) local.phone = msgPhoneEmpty;
    if (form.height.trim() && parseNumber(form.height) === null) local.height = msgNotANumber;
    setLocalErrors(local);
    if (Object.keys(local).length > 0) {
      save.reset();
      return;
    }
    save.mutate(toInput(form));
  };

  return (
    <Screen>
      <Header>
        <RoundButton label="بستن" onClick={dirty ? discard.show : () => goBack()}>
          <XIcon />
        </RoundButton>
        <HeaderTitle title={athlete ? "ویرایش شاگرد" : "شاگرد جدید"} />
        <PillButton onClick={() => onSubmit()} disabled={save.isPending}>
          ذخیره
        </PillButton>
      </Header>
      <ScreenContent>
        <form className={s.form} onSubmit={onSubmit} onKeyDown={focusNextOnEnter} noValidate>
          <div className={s.two}>
            <TextField
              label="نام"
              value={form.firstName}
              onChange={(e) => set("firstName", e.target.value)}
              error={errors.first_name}
              autoComplete="off"
            />
            <TextField
              label="نام خانوادگی"
              value={form.lastName}
              onChange={(e) => set("lastName", e.target.value)}
              error={errors.last_name}
              autoComplete="off"
            />
          </div>
          <TextField
            label="موبایل"
            type="tel"
            inputMode="numeric"
            ltr
            value={form.phone}
            onChange={(e) => set("phone", limitPhoneInput(e.target.value))}
            error={errors.phone}
            autoComplete="off"
          />
          <div className={f.fl}>
            <span className={f.label}>نوع</span>
            <Segmented
              field
              label="نوع"
              options={(["private", "public"] as const).map((t) => ({
                value: t,
                label: athleteTypeLabels[t],
              }))}
              value={form.type}
              onChange={(t) => t && set("type", t)}
            />
            {errors.athlete_type && <span className={f.error}>{errors.athlete_type}</span>}
          </div>
          <div className={f.fl}>
            <span className={f.label}>
              سطح <i className={f.optional}>· اختیاری</i>
            </span>
            <Segmented
              field
              clearable
              label="سطح"
              options={(["beginner", "intermediate", "advanced"] as const).map((l) => ({
                value: l,
                label: experienceLabels[l],
              }))}
              value={form.level}
              onChange={(l) => set("level", l)}
            />
            {errors.experience_level && <span className={f.error}>{errors.experience_level}</span>}
          </div>
          <div className={s.two}>
            <div className={s.wide}>
              <TextField
                label="هدف"
                optional
                value={form.goal}
                onChange={(e) => set("goal", e.target.value)}
                error={errors.goal}
              />
            </div>
            <TextField
              label="قد"
              optional
              inputMode="decimal"
              value={form.height}
              onChange={(e) => set("height", e.target.value)}
              error={errors.height}
              trailing={<span className={f.unit}>سانتی‌متر</span>}
            />
          </div>
          <TextAreaField
            label="آسیب‌دیدگی"
            optional
            placeholder="مثلاً درد زانو، دیسک کمر…"
            value={form.injuries}
            onChange={(e) => set("injuries", e.target.value)}
            error={errors.injuries}
          />
          {athlete && (
            <div className={s.del}>
              <BigButton outline onClick={remove.show}>
                <TrashIcon /> حذف شاگرد
              </BigButton>
            </div>
          )}
        </form>
      </ScreenContent>
      <ConfirmSheet
        open={discard.open}
        onClose={discard.close}
        title="تغییرات ذخیره نشده‌اند"
        text="اگر بیرون بروید، تغییراتی که دادید از بین می‌روند."
        confirmLabel="بیرون رفتن بدون ذخیره"
        onConfirm={() => goBack(1)}
      />
      {athlete && (
        <DeleteAthleteSheet
          open={remove.open}
          onClose={remove.close}
          athlete={athlete}
          onDeleted={() => navigate(paths.home, { replace: true })}
        />
      )}
    </Screen>
  );
}
