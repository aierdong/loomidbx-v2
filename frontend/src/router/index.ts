import { createRouter, createWebHashHistory } from "vue-router";
import BootstrapPage from "@/pages/BootstrapPage.vue";

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "bootstrap",
      component: BootstrapPage,
    },
  ],
});
