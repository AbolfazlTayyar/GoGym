import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useParams, useSearchParams } from "react-router";
import { getAthlete } from "../api/athletes";
import { queryKeys } from "../api/keys";
import { listMeasurements } from "../api/measurements";
import { getPlan, listAthletePlans } from "../api/plans";
import type { Athlete, Measurement, PlanSummary } from "../api/types";
import { BigButton, BottomBar } from "../components/BigButton";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { AlertIcon, BackIcon, ChevLIcon, EditIcon, PhoneIcon } from "../components/icons";
import { MeasurementSheet } from "../components/MeasurementSheet";
import { PillTabs } from "../components/PillTabs";
import { PlanMini } from "../components/PlanMini";
import { DeletePlanSheet, PlanSheet } from "../components/PlanSheet";
import { Screen, ScreenContent } from "../components/Screen";
import { SwipeRow } from "../components/SwipeRow";
import { useSheet } from "../components/useSheet";
import { formatPhone, toPersianDigits } from "../lib/digits";
import { formatJalali, todayISO } from "../lib/jalali";
import { athleteTypeLabels, experienceLabels } from "../lib/labels";
import { paths } from "../lib/paths";
import { planSizeText } from "../lib/planSize";
import { useGoBack } from "../lib/useGoBack";
import { MeasurementsTab } from "./MeasurementsTab";
import s from "./ProfilePage.module.css";

type Tab = "plans" | "measurements";
const tabParam = "tab";

export function ProfilePage() {
  const { id = "" } = useParams();
  const [params, setParams] = useSearchParams();
  // The tab is in the URL, so back from a plan returns to it.
  const tab: Tab = params.get(tabParam) === "measurements" ? "measurements" : "plans";
  const goBack = useGoBack(paths.home);
  const measurementSheet = useSheet("measurement");
  const [editing, setEditing] = useState<Measurement | null>(null);
  const openMeasurement = (m: Measurement | null) => {
    setEditing(m);
    measurementSheet.show();
  };

  const athlete = useQuery({
    queryKey: queryKeys.athletes("detail", id),
    queryFn: () => getAthlete(id),
  });
  const plans = useQuery({
    queryKey: queryKeys.plans("athlete", id),
    queryFn: () => listAthletePlans(id),
  });
  const measurements = useQuery({
    queryKey: queryKeys.measurements("athlete", id),
    queryFn: () => listMeasurements(id),
  });

  const failed = athlete.error ?? plans.error ?? measurements.error;
  if (failed) {
    return (
      <Screen>
        <Header>
          <RoundButton label="بازگشت" onClick={() => goBack()}>
            <BackIcon />
          </RoundButton>
          <HeaderTitle title="شاگرد" />
        </Header>
        <EmptyState
          title={failed.message}
          action={<BigButton to={paths.home}>بازگشت به شاگردها</BigButton>}
        />
      </Screen>
    );
  }
  if (!athlete.data) return <Screen>{null}</Screen>;
  const a = athlete.data;

  return (
    <Screen>
      <Header>
        <RoundButton label="بازگشت" onClick={() => goBack()}>
          <BackIcon />
        </RoundButton>
        <HeaderTitle
          title={`${a.first_name} ${a.last_name}`}
          subtitle={
            <a href={`tel:${a.phone}`} className={s.ph}>
              <PhoneIcon />
              {formatPhone(a.phone)}
            </a>
          }
        />
        <RoundButton label="ویرایش شاگرد" pink to={paths.editAthlete(a.id)}>
          <EditIcon />
        </RoundButton>
      </Header>
      <ScreenContent>
        <div className={s.body}>
          <InfoCard athlete={a} />
          <PillTabs
            options={[
              { value: "plans", label: `برنامه‌ها ${toPersianDigits(plans.data?.length ?? 0)}` },
              {
                value: "measurements",
                label: `اندازه‌ها ${toPersianDigits(measurements.data?.length ?? 0)}`,
              },
            ]}
            value={tab}
            onChange={(t) => setParams({ [tabParam]: t }, { replace: true })}
          />
          {tab === "plans" ? (
            <PlansTab plans={plans.data} />
          ) : (
            <MeasurementsTab measurements={measurements.data ?? []} onOpen={openMeasurement} />
          )}
        </div>
      </ScreenContent>
      <BottomBar>
        {tab === "plans" ? (
          <BigButton to={paths.newPlan(a.id)}>+ برنامه جدید</BigButton>
        ) : (
          <BigButton onClick={() => openMeasurement(null)}>+ ثبت اندازه</BigButton>
        )}
      </BottomBar>
      <MeasurementSheet
        open={measurementSheet.open}
        onClose={measurementSheet.close}
        athleteId={a.id}
        measurements={measurements.data ?? []}
        editing={editing}
      />
    </Screen>
  );
}

// What the coach checks before writing a plan: type, level, height, goal, and injuries in yellow.
function InfoCard({ athlete }: { athlete: Athlete }) {
  return (
    <div className={s.info}>
      <div className={s.chips}>
        <span className={s.pk}>{athleteTypeLabels[athlete.athlete_type]}</span>
        {athlete.experience_level && <span>{experienceLabels[athlete.experience_level]}</span>}
        {athlete.height !== null && <span>قد {toPersianDigits(athlete.height)}</span>}
        {athlete.goal && <span>هدف: {athlete.goal}</span>}
      </div>
      {athlete.injuries && (
        <div className={s.inj}>
          <AlertIcon />
          آسیب‌دیدگی: {athlete.injuries}
        </div>
      )}
    </div>
  );
}

function PlansTab({ plans }: { plans: PlanSummary[] | undefined }) {
  const current = plans?.find((p) => p.is_current) ?? null;
  const currentDetail = useQuery({
    queryKey: queryKeys.plans("detail", current?.id),
    queryFn: () => getPlan(current!.id),
    enabled: current !== null,
  });
  const queryClient = useQueryClient();
  // One row open at a time, as in iOS: swiping another closes this one.
  const [swipedId, setSwipedId] = useState<string | null>(null);
  // The plan a row's edit or delete was tapped on, kept after its sheet closes so the sheet slides down whole.
  const [target, setTarget] = useState<PlanSummary | null>(null);
  const editSheet = useSheet("plan");
  const deleteSheet = useSheet("delete-plan");

  const swipe = (p: PlanSummary) => ({
    open: swipedId === p.id,
    onOpenChange: (open: boolean) => setSwipedId(open ? p.id : null),
    onEdit: () => {
      setSwipedId(null);
      setTarget(p);
      editSheet.show();
    },
    onDelete: () => {
      setSwipedId(null);
      setTarget(p);
      deleteSheet.show();
    },
  });

  if (!plans) return <p className={s.note}>در حال بارگذاری…</p>;
  if (plans.length === 0) {
    return (
      <EmptyState title="هنوز برنامه‌ای ندارد" text="با + برنامه جدید اولین برنامه را بنویسید." />
    );
  }

  // Newest first: plans that haven't started, the current one, then the older ones.
  const today = todayISO();
  const upcoming = plans.filter((p) => !p.is_current && p.start_date > today);
  const older = plans.filter((p) => !p.is_current && p.start_date <= today);

  return (
    <>
      <div className={s.plist}>
        {upcoming.map((p) => (
          <SwipeRow key={p.id} {...swipe(p)}>
            <Link to={paths.dayView(p.id)} className={s.pc}>
              <div className={s.top}>
                <b>{p.title}</b>
                <span className={`${s.badge} ${s.soon}`}>بعدی</span>
              </div>
              <span className={s.sub}>شروع {formatJalali(p.start_date)}</span>
            </Link>
          </SwipeRow>
        ))}
        {current && (
          <SwipeRow {...swipe(current)}>
            <div className={`${s.pc} ${s.cur}`}>
              <Link to={paths.dayView(current.id)}>
                <div className={s.top}>
                  <b>{current.title}</b>
                  <span className={`${s.badge} ${s.now}`}>فعلی</span>
                </div>
                <span className={s.sub}>
                  شروع {formatJalali(current.start_date)} · {planSizeText(current)}
                </span>
              </Link>
              {currentDetail.data && (
                <PlanMini
                  days={currentDetail.data.days}
                  dayLink={(label) => paths.dayView(current.id, label)}
                />
              )}
            </div>
          </SwipeRow>
        )}
      </div>
      {older.length > 0 && (
        <>
          <div className={s.lbl}>برنامه‌های قبلی</div>
          <div className={s.hist}>
            {older.map((p) => (
              <SwipeRow key={p.id} {...swipe(p)} className={s.hrow}>
                <Link to={paths.dayView(p.id)} className={s.hr}>
                  <div className={s.nm}>
                    <b>{p.title}</b>
                    <span>
                      شروع {formatJalali(p.start_date)} · {planSizeText(p)}
                    </span>
                  </div>
                  <span className={s.go}>
                    <ChevLIcon />
                  </span>
                </Link>
              </SwipeRow>
            ))}
          </div>
        </>
      )}
      {older.length === 0 && <div className={s.space} />}
      {target && (
        <>
          <PlanSheet
            open={editSheet.open}
            onClose={editSheet.close}
            plan={target}
            onSaved={() => {
              queryClient.invalidateQueries({ queryKey: queryKeys.plans() });
              editSheet.close();
            }}
          />
          <DeletePlanSheet
            key={target.id}
            open={deleteSheet.open}
            onClose={deleteSheet.close}
            plan={target}
            onDeleted={() => {
              queryClient.removeQueries({ queryKey: queryKeys.plans("detail", target.id) });
              queryClient.invalidateQueries({ queryKey: queryKeys.plans() });
              deleteSheet.close();
            }}
          />
        </>
      )}
    </>
  );
}
