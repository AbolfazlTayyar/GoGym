import { createBrowserRouter } from "react-router";
import { RequireAuth } from "./components/RequireAuth";
import { TabLayout } from "./components/TabLayout";
import { EditAthletePage, NewAthletePage } from "./pages/AthleteFormPage";
import { HomePage } from "./pages/HomePage";
import { LoginPage } from "./pages/LoginPage";
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
      { path: "/athletes/new", element: <NewAthletePage /> },
      { path: "/athletes/:id/edit", element: <EditAthletePage /> },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
