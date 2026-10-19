import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, KickerButton, RoundButton } from "../components/Header";
import { PlusIcon, SparkIcon, UsersIcon } from "../components/icons";
import { ScreenContent } from "../components/Screen";
import { SignOutSheet } from "../components/SignOutSheet";
import { useSheet } from "../components/useSheet";

export function AthletesPage() {
  const account = useSheet("account");

  return (
    <>
      <Header>
        <HeaderTitle
          kicker={
            <KickerButton onClick={account.show}>
              <SparkIcon /> سلام مربی
            </KickerButton>
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
      <SignOutSheet open={account.open} onClose={account.close} />
    </>
  );
}
