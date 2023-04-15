import type { RouteDetail, RoutePrecondition } from "svelte-spa-router";
import { wrap } from "svelte-spa-router/wrap";

import config from "@/config";

import Login from "@/pages/Login.svelte";
import Loading from "@/pages/Loading.svelte";

import { UserStore } from "@/lib/auth";

const baseConditions: RoutePrecondition[] = [
  async (_: RouteDetail) => {
    await config.refresh();
    return true;
  },
  async (_: RouteDetail) => {
    await UserStore.refresh();
    return true;
  },
];

const isAuthenticatedCondition = (shouldBe: boolean = true) => {
  return async (_: RouteDetail) => {
    return UserStore.isAvailable() == shouldBe;
  };
};

const processLogOut = () => {
  return async () => {
    // At this point, the user passed the authenticated condition, so we know
    // for sure that he/she is authenticated

    await UserStore.clear();

    return false;
  };
};

const routes = {
  "/": wrap({
    asyncComponent: () => import("@/pages/Dashboard.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
  "/login": wrap({
    component: Login,
    conditions: baseConditions.concat([isAuthenticatedCondition(false)]),
    userData: {
      onFailureRedirectTo: "/",
    },
  }),
  "/logout": wrap({
    component: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition(), processLogOut()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
  "/settings": wrap({
    asyncComponent: () => import("@/pages/Settings.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
  "/modules/:namespace/:name/:provider/:version?": wrap({
    asyncComponent: () => import("@/pages/Artifact.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
  "/providers/:namespace/:name/:version?": wrap({
    asyncComponent: () => import("@/pages/Artifact.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
  "*": wrap({
    asyncComponent: () => import("@/pages/Error404.svelte"),
    loadingComponent: Loading,
    conditions: baseConditions.concat([isAuthenticatedCondition()]),
    userData: {
      onFailureRedirectTo: "/login",
    },
  }),
};

export default routes;