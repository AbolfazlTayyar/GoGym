import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { DumbbellIcon, PlusIcon } from "../components/icons";
import { ScreenContent } from "../components/Screen";

export function MovementsPage() {
  return (
    <>
      <Header>
        <HeaderTitle kicker="کتابخانه حرکات" title="حرکات" />
        <RoundButton label="حرکت جدید" pink>
          <PlusIcon />
        </RoundButton>
      </Header>
      <ScreenContent>
        <EmptyState icon={<DumbbellIcon />} title="هنوز حرکتی نیست" />
      </ScreenContent>
    </>
  );
}
