import { Outlet } from "react-router";
import { BottomNav } from "./BottomNav";
import { Screen } from "./Screen";

// The two top-level pages share this frame: each renders its header and content, the bottom bar stays.
export function TabLayout() {
  return (
    <Screen>
      <Outlet />
      <BottomNav />
    </Screen>
  );
}
