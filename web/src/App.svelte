<script lang="ts">
  import Router, {
    push,
    replace,
    type ConditionsFailedEvent,
    type RouteLoadingEvent,
  } from "svelte-spa-router";

  import routes from "./routes";

  let title: string = "Terralist";

  let currentRoute: string = undefined;

  const onRouteLoading = (e: RouteLoadingEvent) => {
    if (e?.detail?.location === currentRoute) {
      return;
    }

    if (currentRoute) {
      push(e?.detail?.location);
    }

    currentRoute = e?.detail?.location;
  };

  const onRouteFailure = (e: ConditionsFailedEvent) => {
    if (e.detail?.userData) {
      replace(e.detail.userData["onFailureRedirectTo"]);
    }
  };
</script>

<svelte:head>
  <title>{title}</title>
</svelte:head>

<main>
  <Router
    {routes}
    on:routeLoading={onRouteLoading}
    on:conditionsFailed={onRouteFailure}
  />
  <!-- <Header /> -->
  <!-- <Dashboard /> -->
</main>
