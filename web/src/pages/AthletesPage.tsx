import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { PlusIcon, SparkIcon, UsersIcon } from "../components/icons";
import { ScreenContent } from "../components/Screen";

export function AthletesPage() {
  return (
    <>
      <Header>
        <HeaderTitle
          kicker={
            <>
              <SparkIcon /> سلام مربی
            </>
          }
          title="شاگردهای من"
        />
        <RoundButton label="شاگرد جدید" pink to="/athletes/new">
          <PlusIcon />
        </RoundButton>
      </Header>
      <ScreenContent>
        <EmptyState
          icon={<UsersIcon />}
          title="هنوز شاگردی ندارید"
          text="با دکمه + اولین شاگرد را اضافه کنید."
        />
      </ScreenContent>
    </>
  );
}
