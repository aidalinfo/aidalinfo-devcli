import { createRouter, createWebHashHistory, RouteRecordRaw } from "vue-router";

// Pages
import IndexPage from "@/pages/Index.vue";
// Ajoute ici d'autres pages si besoin
import SetupPage from "@/pages/Setup.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "Home",
    component: IndexPage,
  },
  {
    path: "/settings",
    name: "Settings",
    component: () => import("@/pages/Settings.vue"),
  },
  {
    path: "/setup",
    name: "Setup",
    component: SetupPage,
  },
  // Ajoute d'autres routes ici
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
