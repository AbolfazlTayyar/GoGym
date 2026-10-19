import { useQuery } from "@tanstack/react-query";
import { useParams, useSearchParams } from "react-router";
import { getAthlete } from "../api/athletes";
import { queryKeys } from "../api/keys";
import { getPlan } from "../api/plans";
import { BigButton } from "../components/BigButton";
import { BlockLine, DayHeading } from "../components/BlockLine";
import { DayPills } from "../components/DayPills";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { BackIcon, EditIcon, UserIcon } from "../components/icons";
import { Screen } from "../components/Screen";
import { formatJalali } from "../lib/jalali";
import { paths } from "../lib/paths";
import { useGoBack } from "../lib/useGoBack";
import s from "./DayViewPage.module.css";

const dayParam = "day";

export function DayViewPage() {
  const { planId = "" } = useParams();
  const [params, setParams] = useSearchParams();
  const goBack = useGoBack(paths.home);

  const plan = useQuery({
    queryKey: queryKeys.plans("detail", planId),
    queryFn: () => getPlan(planId),
  });
  const athleteId = plan.data?.athlete_id ?? "";
  const athlete = useQuery({
    queryKey: queryKeys.athletes("detail", athleteId),
    queryFn: () => getAthlete(athleteId),
    enabled: athleteId !== "",
  });

  const failed = plan.error ?? athlete.error;
  if (failed) {
    return (
      <Screen>
        <Header>
          <RoundButton label="بازگشت" onClick={() => goBack()}>
            <BackIcon />
          </RoundButton>
          <HeaderTitle title="برنامه" />
        </Header>
        <EmptyState
          title={failed.message}
          action={<BigButton to={paths.home}>بازگشت به شاگردها</BigButton>}
        />
      </Screen>
    );
  }

  const days = plan.data?.days ?? [];
  // An unknown or missing day opens the first one.
  const day = days.find((d) => d.label === params.get(dayParam)) ?? days[0];

  return (
    <Screen>
      <Header>
        <RoundButton label="بازگشت" onClick={() => goBack()}>
          <BackIcon />
        </RoundButton>
        <HeaderTitle
          title={athlete.data ? `${athlete.data.first_name} ${athlete.data.last_name}` : "…"}
          subtitle={
            plan.data ? `${plan.data.title} · شروع ${formatJalali(plan.data.start_date)}` : " "
          }
        />
        {athleteId && (
          <RoundButton label="پروفایل شاگرد" to={paths.profile(athleteId)}>
            <UserIcon />
          </RoundButton>
        )}
        <RoundButton label="ویرایش برنامه" pink to={paths.builder(planId, day?.label)}>
          <EditIcon />
        </RoundButton>
      </Header>
      <DayPills
        labels={days.map((d) => d.label)}
        current={day?.label ?? null}
        // Replace, not push: back leaves the plan in one step, whichever days were looked at.
        onSelect={(label) => setParams({ [dayParam]: label }, { replace: true })}
      />
      <div className={s.sheet}>
        {plan.isPending ? (
          <p className={s.note}>در حال بارگذاری…</p>
        ) : !day ? (
          <p className={s.note}>این برنامه هنوز روزی ندارد.</p>
        ) : (
          <>
            <DayHeading label={day.label} lines={day.blocks.length} />
            {day.blocks.length === 0 && <p className={s.note}>این روز هنوز حرکتی ندارد.</p>}
            {day.blocks.map((block, i) => (
              <BlockLine key={block.id} index={i} block={block} />
            ))}
          </>
        )}
      </div>
    </Screen>
  );
}
