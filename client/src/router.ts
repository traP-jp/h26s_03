import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/todo",
      name: "todo",
      component: () => import("./views/TodoView.vue"),
    },
    {
      path: "/",
      name: "dashboard",
      component: () => import("./views/DashboardView.vue"),
    },
  ],
});

export default router;
