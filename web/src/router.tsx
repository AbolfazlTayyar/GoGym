import { createBrowserRouter } from "react-router";
import { RequireAuth } from "./components/RequireAuth";
import { TabLayout } from "./components/TabLayout";
import { AthletesPage } from "./pages/AthletesPage";
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
          { path: "/", element: <AthletesPage /> },
          { path: "/movements", element: <MovementsPage /> },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
