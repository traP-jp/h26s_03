const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

import createClient from "openapi-fetch";

import type { components, paths } from "../gen/api-types";

export type Task = components["schemas"]["Task"];

const client = createClient<paths>({
  baseUrl: apiBase,
});

function raiseApiError(error: unknown): never {
  if (error instanceof Error) {
    throw error;
  }
  throw new Error("request failed");
}

export async function initializeData(): Promise<void> {
  const { error } = await client.POST("/api/initialize");
  if (error) {
    raiseApiError(error);
  }
}

export async function getTasks(): Promise<Task[]> {
  const { data, error } = await client.GET("/api/tasks");
  if (error || !data) {
    raiseApiError(error);
  }
  return data.data;
}

export async function addTask(title: string): Promise<void> {
  const { error } = await client.POST("/api/tasks", {
    body: { title },
  });
  if (error) {
    raiseApiError(error);
  }
}
