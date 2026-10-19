import "@fontsource-variable/vazirmatn";
import "@fontsource-variable/inter";
import "./styles/tokens.css";
import "./styles/base.css";
import { QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router/dom";
import { onUnauthorized } from "./api/client";
import { queryClient } from "./api/queryClient";
import { goToSignIn } from "./api/session";
import { router } from "./router";

onUnauthorized(() =>
  goToSignIn(router.navigate, queryClient, window.location.pathname + window.location.search),
);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
