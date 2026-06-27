const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

import createClient from "openapi-fetch";

import type { paths } from "../gen/api-types";

const client = createClient<paths>({
  baseUrl: apiBase,
});

const raiseApiError = (error: unknown): never => {
  if (error instanceof Error) {
    throw error;
  }
  throw new Error("request failed");
};

export const initializeData = async (): Promise<void> => {
  const { error } = await client.POST("/api/initialize");
  if (error) {
    raiseApiError(error);
  }
};
