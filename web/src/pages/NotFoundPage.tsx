import { Link } from "react-router";
import { EmptyState } from "../components/EmptyState";
import { Header, HeaderTitle, RoundButton } from "../components/Header";
import { BackIcon } from "../components/icons";
import { Screen, ScreenContent } from "../components/Screen";

export function NotFoundPage() {
  return (
    <Screen>
      <Header>
        <RoundButton label="خانه" to="/">
          <BackIcon />
        </RoundButton>
        <HeaderTitle title="پیدا نشد" />
      </Header>
      <ScreenContent>
        <EmptyState
          title="این صفحه وجود ندارد"
          text="شاید حذف شده باشد."
          action={<Link to="/">بازگشت به شاگردها</Link>}
        />
      </ScreenContent>
    </Screen>
  );
}
