import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "dashboard",
      component: () => import("./views/DashboardView.vue"),
    },
    {
      path: "/members",
      name: "members",
      component: () => import("./views/MembersView.vue"),
    },
  ],
});

export default router;
