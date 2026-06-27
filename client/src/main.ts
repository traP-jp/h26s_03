import { createApp } from "vue";

import App from "./App.vue";
import router from "./router";

import "./style.css";

const enableMocking = async () => {
  if (!import.meta.env.DEV || import.meta.env.VITE_API_MOCKING === "false") {
    return;
  }

  const { worker } = await import("./mocks/browser");

  return worker.start({
    onUnhandledRequest: "bypass",
  });
};

void enableMocking().then(() => {
  createApp(App).use(router).mount("#app");
});
