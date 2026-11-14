import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router";
import { ApiError } from "../api/client";
import { ErrorCode } from "../api/codes";
import { queryKeys } from "../api/keys";
import { getSharedPlan } from "../api/publicPlans";
import type { SharedPlan } from "../api/types";
import { BigButton, BottomBar } from "../components/BigButton";
import { BlockLine, DayHeading } from "../components/BlockLine";
import { EmptyState } from "../components/EmptyState";
import { PhoneIcon } from "../components/icons";
import { Screen, ScreenContent } from "../components/Screen";
import { useToast } from "../components/Toast";
import { formatPhone } from "../lib/digits";
import { formatJalali } from "../lib/jalali";
import { downloadPlanPdf, planPdfFileName } from "../lib/planPdf";
import s from "./SharedPlanPage.module.css";

const msgPdfFailed = "ساخت PDF نشد؛ دوباره امتحان کنید.";

// The athlete's view of a plan, opened from the link the coach sent; no login, no app frame.
export function SharedPlanPage() {
  const { token = "" } = useParams();
  const shared = useQuery({
    queryKey: queryKeys.sharedPlan(token),
    queryFn: () => getSharedPlan(token),
    // A 404 is the answer, not a hiccup: asking again won't change it.
    retry: (count, err) => !(err instanceof ApiError && err.status === 404) && count < 2,
  });
  useNoIndex();

  if (shared.isPending) {
    return (
      <Screen>
        <p className={s.note}>در حال بارگذاری…</p>
      </Screen>
    );
  }

  if (shared.isError) {
    const gone = shared.error instanceof ApiError && shared.error.code === ErrorCode.notFound;
    return (
      <Screen>
        <div className={s.alone}>
          <EmptyState
            title={gone ? "این لینک دیگر معتبر نیست" : shared.error.message}
            text={gone ? "برای لینک جدید به مربی خود پیام دهید." : undefined}
            action={
              gone ? undefined : <BigButton onClick={() => shared.refetch()}>دوباره</BigButton>
            }
          />
        </div>
      </Screen>
    );
  }

  const { coach, athlete, plan } = shared.data;
  if (!athlete || !plan) {
    return (
      <Screen>
        <ScreenContent>
          <div className={s.page}>
            <CoachCard coach={coach} />
            <EmptyState
              title="این لینک منقضی شده است"
              text="برای لینک جدید به مربی خود پیام دهید."
            />
          </div>
        </ScreenContent>
      </Screen>
    );
  }

  return <ActivePlan coach={coach} athlete={athlete} plan={plan} />;
}

type ActivePlanProps = {
  coach: SharedPlan["coach"];
  athlete: NonNullable<SharedPlan["athlete"]>;
  plan: NonNullable<SharedPlan["plan"]>;
};

function ActivePlan({ coach, athlete, plan }: ActivePlanProps) {
  const content = useRef<HTMLDivElement>(null);
  const [saving, setSaving] = useState(false);
  const { toast, show } = useToast();

  async function savePdf() {
    if (!content.current) return;
    setSaving(true);
    try {
      await downloadPlanPdf(
        content.current,
        planPdfFileName(athlete.first_name, athlete.last_name, plan.title),
      );
    } catch {
      show(msgPdfFailed);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Screen>
      <ScreenContent>
        {/* Everything in here is what the PDF shows. */}
        <div ref={content} className={s.page}>
          <CoachCard coach={coach} />
          <div className={s.head}>
            <h1>
              {athlete.first_name} {athlete.last_name}
            </h1>
            <span>
              {plan.title} · شروع {formatJalali(plan.start_date)}
            </span>
          </div>
          {plan.note && <p className={s.planNote}>{plan.note}</p>}
          {plan.days.length === 0 && <p className={s.note}>این برنامه هنوز روزی ندارد.</p>}
          {plan.days.map((day) => (
            <section key={day.id} className={s.sheet}>
              <DayHeading label={day.label} lines={day.blocks.length} />
              {day.blocks.map((block, i) => (
                <BlockLine key={block.id} index={i} block={block} />
              ))}
            </section>
          ))}
          <p className={s.footer}>ساخته‌شده با gogymapp</p>
        </div>
      </ScreenContent>
      <BottomBar>
        <BigButton onClick={() => void savePdf()} disabled={saving}>
          {saving ? "در حال ساختن PDF…" : "دریافت PDF"}
        </BigButton>
      </BottomBar>
      {toast}
    </Screen>
  );
}

function CoachCard({ coach }: { coach: SharedPlan["coach"] }) {
  return (
    <div className={s.coach}>
      <div className={s.who}>
        <span>مربی</span>
        <b>
          {coach.first_name} {coach.last_name}
        </b>
      </div>
      <a className={s.phone} href={`tel:${coach.phone}`}>
        <PhoneIcon />
        <span dir="ltr">{formatPhone(coach.phone)}</span>
      </a>
    </div>
  );
}

// A plan's page is for the athlete who has the link, not for search engines.
function useNoIndex() {
  useEffect(() => {
    const meta = document.createElement("meta");
    meta.name = "robots";
    meta.content = "noindex";
    document.head.append(meta);
    return () => meta.remove();
  }, []);
}
