import { createOpenApiHttp } from "openapi-msw";

import type { paths } from "../gen/api-types";

const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";
const http = createOpenApiHttp<paths>({ baseUrl: apiBase });

export const handlers = [
  http.post("/api/initialize", ({ response }) => {
    return response(204).empty();
  }),
];
