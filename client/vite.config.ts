import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite-plus";

export default defineConfig({
  plugins: [vue()],
  fmt: {},
  lint: { options: { typeAware: true, typeCheck: true } },
});
