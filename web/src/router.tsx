import { createBrowserRouter } from "react-router";
import { TabLayout } from "./components/TabLayout";
import { AthletesPage } from "./pages/AthletesPage";
import { MovementsPage } from "./pages/MovementsPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { Placeholder } from "./Placeholder";

export const router = createBrowserRouter([
  {
    element: <TabLayout />,
    children: [
      { path: "/", element: <AthletesPage /> },
      { path: "/movements", element: <MovementsPage /> },
    ],
  },
  { path: "/status", element: <Placeholder /> },
  { path: "*", element: <NotFoundPage /> },
]);
