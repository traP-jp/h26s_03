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
      path: "/polls/create",
      name: "createPoll",
      component: () => import("./views/CreatePollView.vue"),
    }
  ],
});

export default router;
