import { replace, type RouteDetail, type RoutePrecondition } from "svelte-spa-router";
import { wrap } from "svelte-spa-router/wrap";

import Login from "@/pages/Login.svelte";
import Loading from "@/pages/Loading.svelte";

const baseConditions: RoutePrecondition[] = [
  // async (d: RouteDetail) => {
  //   return true;
  // },
];

const isAuthenticatedCondition = (shouldBe: boolean = true) => {
  return async (_: RouteDetail) => {
    // TODO: If not authenticated, should be redirected back to /login
    return true;
  };
};

const processLogOut = () => {
  return async () => {
    // At this point, the user passed the authenticated condition, so we know
    // for sure that he/she is authenticated

    // Clear store

    // Redirect back to login
    // replace("/login");

    setTimeout(() => replace("/login"), 5000);

    return true;
  };
};

const routes = {
  "/": wrap({
    asyncComponent: () => import("@/pages/Dashboard.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
  }),
  "/login": wrap({
    component: Login,
    conditions: baseConditions.concat([isAuthenticatedCondition(false)]),
  }),
  "/logout": wrap({
    component: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition(), processLogOut()]),
  }),
  "/settings": wrap({
    asyncComponent: () => import("@/pages/Settings.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
  }),
  "/:namespace/:name/:provider?/:version": wrap({
    asyncComponent: () => import("@/pages/Artifact.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
  }),
  "*": wrap({
    asyncComponent: () => import("@/pages/Error404.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
  }),
};

export default routes;