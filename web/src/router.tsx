import { createBrowserRouter } from "react-router";
import { RequireAuth } from "./components/RequireAuth";
import { TabLayout } from "./components/TabLayout";
import { EditAthletePage, NewAthletePage } from "./pages/AthleteFormPage";
import { BuilderPage } from "./pages/builder/BuilderPage";
import { DayViewPage } from "./pages/DayViewPage";
import { HomePage } from "./pages/HomePage";
import { LoginPage } from "./pages/LoginPage";
import { NewPlanPage } from "./pages/NewPlanPage";
import { MovementsPage } from "./pages/MovementsPage";
import { NotFoundPage } from "./pages/NotFoundPage";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <TabLayout />,
        children: [
          { path: "/", element: <HomePage /> },
          { path: "/movements", element: <MovementsPage /> },
        ],
      },
      { path: "/plans/:planId", element: <DayViewPage /> },
      { path: "/plans/:planId/edit", element: <BuilderPage /> },
      { path: "/athletes/new", element: <NewAthletePage /> },
      { path: "/athletes/:id/edit", element: <EditAthletePage /> },
      { path: "/athletes/:id/plans/new", element: <NewPlanPage /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
