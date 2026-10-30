import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef, useState, type ReactNode } from "react";
import { useNavigate, useParams } from "react-router";
import { getAthlete } from "../api/athletes";
import { fieldErrors } from "../api/client";
import { queryKeys } from "../api/keys";
import { addDay, createPlan, getPlan, listAthletePlans } from "../api/plans";
import type { Athlete, PlanDetail, PlanSummary } from "../api/types";
import { BigButton, BottomBar } from "../components/BigButton";
import { CopyPickerSheet } from "../components/CopyPickerSheet";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { ChevLIcon, XIcon } from "../components/icons";
import { JalaliDateField } from "../components/JalaliDateField";
import { PlanMini } from "../components/PlanMini";
import { Screen, ScreenContent } from "../components/Screen";
import { TextField } from "../components/TextField";
import { useSheet } from "../components/useSheet";
import { msgTitleEmpty } from "../lib/formMessages";
import { planTitleFor, todayISO } from "../lib/jalali";
import { paths } from "../lib/paths";
import { planSizeText } from "../lib/planSize";
import { useGoBack } from "../lib/useGoBack";
import s from "./NewPlanPage.module.css";

type Choice = "latest" | "other" | "empty";

// An empty plan gets day A, so the builder has a day to open.
const firstDayLabel = "A";

export function NewPlanPage() {
  const { id: athleteId = "" } = useParams();
  const athlete = useQuery({
    queryKey: queryKeys.athletes("detail", athleteId),
    queryFn: () => getAthlete(athleteId),
  });
  const plans = useQuery({
    queryKey: queryKeys.plans("athlete", athleteId),
    queryFn: () => listAthletePlans(athleteId),
  });

  if (athlete.isError || plans.isError) {
    return (
      <Screen>
        <EmptyState
          title={(athlete.error ?? plans.error)!.message}
          action={<BigButton to={paths.home}>بازگشت به شاگردها</BigButton>}
        />
      </Screen>
    );
  }
  if (!athlete.data || !plans.data) return <Screen>{null}</Screen>;
  // The athlete's newest plan is the one a new plan is most like.
  return <NewPlanForm athlete={athlete.data} latest={plans.data[0] ?? null} />;
}

function NewPlanForm({ athlete, latest }: { athlete: Athlete; latest: PlanSummary | null }) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const goBack = useGoBack(paths.profile(athlete.id));
  const picker = useSheet("copy-from");
  const [choice, setChoice] = useState<Choice>(latest ? "latest" : "empty");
  const [other, setOther] = useState<{ athlete: Athlete; plan: PlanSummary } | null>(null);
  const [startDate, setStartDate] = useState(() => todayISO());
  const [title, setTitle] = useState(() => planTitleFor(startDate));
  // Caught here so the message can name the field, which the API's "is required" can't.
  const [titleError, setTitleError] = useState<string | null>(null);
  // A retried save reuses the plan the first attempt created, so a failed day A doesn't leave two plans.
  const created = useRef<PlanDetail | null>(null);

  const latestDetail = usePlanDetail(latest?.id);
  const otherDetail = usePlanDetail(other?.plan.id);

  const sourceId =
    choice === "latest" ? latest?.id : choice === "other" ? other?.plan.id : undefined;

  const save = useMutation({
    mutationFn: async () => {
      created.current ??= await createPlan({
        athlete_id: athlete.id,
        title: title.trim(),
        start_date: startDate,
        ...(sourceId ? { source_plan_id: sourceId } : {}),
      });
      const plan = created.current;
      if (plan.days.length === 0) await addDay(plan.id, firstDayLabel, 0);
      return plan;
    },
    onSuccess: async (plan) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.plans() });
      // Replacing this screen, so back from the builder goes to where the coach came from.
      navigate(paths.builder(plan.id), { replace: true });
    },
  });
  const errors = fieldErrors(save.error);

  const submit = () => {
    if (!title.trim()) {
      setTitleError(msgTitleEmpty);
      save.reset();
      return;
    }
    setTitleError(null);
    save.mutate();
  };

  const pickOther = () => {
    if (other) setChoice("other");
    else picker.show();
  };

  return (
    <Screen>
      <Header>
        <RoundButton label="بستن" onClick={() => goBack()}>
          <XIcon />
        </RoundButton>
        <HeaderTitle title="برنامه جدید" subtitle={`${athlete.first_name} ${athlete.last_name}`} />
      </Header>
      <ScreenContent>
        <div className={s.q}>از کجا شروع کنیم؟</div>
        <div className={s.choices} role="radiogroup">
          {latest && (
            <ChoiceCard
              on={choice === "latest"}
              onSelect={() => setChoice("latest")}
              title={`کپی «${latest.title}» ${athlete.first_name}`}
              text={`${planSizeText(latest)} · فقط تغییرات را وارد کنید`}
            >
              {latestDetail && <PlanMini days={latestDetail.days} />}
            </ChoiceCard>
          )}
          <ChoiceCard
            on={choice === "other"}
            onSelect={pickOther}
            title={
              other
                ? `کپی «${other.plan.title}» ${other.athlete.first_name} ${other.athlete.last_name}`
                : "کپی برنامه یک شاگرد دیگر"
            }
            text={other ? planSizeText(other.plan) : "شاگرد و برنامه را انتخاب کنید"}
            trailing={
              <span
                className={s.faint}
                role="button"
                aria-label="انتخاب برنامه دیگر"
                onClick={(e) => {
                  e.stopPropagation();
                  picker.show();
                }}
              >
                <ChevLIcon />
              </span>
            }
          >
            {choice === "other" && otherDetail && <PlanMini days={otherDetail.days} />}
          </ChoiceCard>
          <ChoiceCard
            on={choice === "empty"}
            onSelect={() => setChoice("empty")}
            title="برنامه خالی"
            text="از صفر بنویسید"
          />
          {errors.source_plan_id && <span className={s.error}>{errors.source_plan_id}</span>}
        </div>
        <div className={s.fields}>
          <TextField
            label="عنوان"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            error={titleError ?? errors.title}
          />
          <JalaliDateField
            label="تاریخ شروع"
            value={startDate}
            onChange={setStartDate}
            error={errors.start_date}
          />
        </div>
      </ScreenContent>
      <BottomBar>
        <BigButton onClick={submit} disabled={save.isPending}>
          ادامه
        </BigButton>
        {save.isError && Object.keys(errors).length === 0 && (
          <p className={s.error} role="alert">
            {save.error.message}
          </p>
        )}
      </BottomBar>
      <CopyPickerSheet
        open={picker.open}
        onClose={picker.close}
        exceptAthleteId={athlete.id}
        onPick={(a, plan) => {
          setOther({ athlete: a, plan });
          setChoice("other");
          picker.close();
        }}
      />
    </Screen>
  );
}

function usePlanDetail(planId: string | undefined) {
  const plan = useQuery({
    queryKey: queryKeys.plans("detail", planId),
    queryFn: () => getPlan(planId!),
    enabled: planId !== undefined,
  });
  return plan.data;
}

type ChoiceCardProps = {
  on: boolean;
  onSelect: () => void;
  title: string;
  text: string;
  trailing?: ReactNode;
  children?: ReactNode;
};

function ChoiceCard({ on, onSelect, title, text, trailing, children }: ChoiceCardProps) {
  return (
    <div
      role="radio"
      aria-checked={on}
      tabIndex={0}
      className={on ? `${s.ch} ${s.on}` : s.ch}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") onSelect();
      }}
    >
      <span className={on ? `${s.radio} ${s.on}` : s.radio} />
      <div className={s.ct}>
        <b>{title}</b>
        <span>{text}</span>
        {children}
      </div>
      {trailing}
    </div>
  );
}
