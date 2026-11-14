import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useParams, useSearchParams } from "react-router";
import { getAthlete } from "../api/athletes";
import { ApiError } from "../api/client";
import { queryKeys } from "../api/keys";
import { getPlan, sharePlan } from "../api/plans";
import { BigButton } from "../components/BigButton";
import { BlockLine, DayHeading } from "../components/BlockLine";
import { DayPills } from "../components/DayPills";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { BackIcon, EditIcon, LinkIcon, UserIcon } from "../components/icons";
import { Screen } from "../components/Screen";
import { useToast } from "../components/Toast";
import { copyPendingText } from "../lib/clipboard";
import { formatJalali } from "../lib/jalali";
import { paths } from "../lib/paths";
import { linkCopiedMessage, msgCopyFailed, sharedPlanURL } from "../lib/shareLink";
import { useGoBack } from "../lib/useGoBack";
import s from "./DayViewPage.module.css";

const dayParam = "day";

export function DayViewPage() {
  const { planId = "" } = useParams();
  const [params, setParams] = useSearchParams();
  const goBack = useGoBack(paths.home);
  const { toast, show } = useToast();
  const [sharing, setSharing] = useState(false);

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

  // Every tap makes a new link and ends the previous one (docs/adr/0035). The copy starts inside the tap,
  // before the request answers, because iOS refuses a copy once the tap is over.
  async function copyShareLink() {
    setSharing(true);
    const link = sharePlan(planId);
    const copied = copyPendingText(link.then((l) => sharedPlanURL(l.token)));
    try {
      const { expires_at } = await link;
      await copied.then(
        () => show(linkCopiedMessage(expires_at)),
        () => show(msgCopyFailed),
      );
    } catch (err) {
      // The request failed, so the copy failed with it; its message is the one that matters.
      copied.catch(() => {});
      show(err instanceof ApiError ? err.message : msgCopyFailed);
    } finally {
      setSharing(false);
    }
  }

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
        <RoundButton
          label="اشتراک‌گذاری لینک"
          onClick={() => void copyShareLink()}
          disabled={sharing || !plan.data}
        >
          <LinkIcon />
        </RoundButton>
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
            {plan.data?.note && <p className={s.planNote}>{plan.data.note}</p>}
            <DayHeading label={day.label} lines={day.blocks.length} />
            {day.blocks.length === 0 && (
              <Link className={s.note} to={paths.builder(planId, day.label)}>
                این روز هنوز حرکتی ندارد. برای افزودن بزنید.
              </Link>
            )}
            {/* Coaches tap a line to edit it, as in the builder, so it opens there with this line open. */}
            {day.blocks.map((block, i) => (
              <Link
                key={block.id}
                className={s.line}
                to={paths.builder(planId, day.label, block.id)}
              >
                <BlockLine index={i} block={block} />
              </Link>
            ))}
          </>
        )}
      </div>
      {toast}
    </Screen>
  );
}
