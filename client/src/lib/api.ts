const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

import createClient from "openapi-fetch";

import type { components, paths } from "../gen/api-types";

export type FeedItem = components["schemas"]["FeedItem"];
export type Member = components["schemas"]["Member"];

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

export async function getFeed(): Promise<FeedItem[]> {
  const { data, error } = await client.GET("/api/feed");
  if (error || !data) {
    raiseApiError(error);
  }
  return data.data;
}

export async function getMembers(): Promise<Member[]> {
  const { data, error } = await client.GET("/api/members");
  if (error || !data) {
    raiseApiError(error);
  }
  return data.data;
}

export async function addTask(title: string, memberId: number): Promise<void> {
  const { error } = await client.POST("/api/tasks", {
    body: { title, member_id: memberId },
  });
  if (error) {
    raiseApiError(error);
  }
}
