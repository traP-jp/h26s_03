import createClient from "openapi-fetch";

import type { components, paths } from "../gen/api-types";

type CreatePollRequest = components["schemas"]["CreatePollRequest"];
type Poll = components["schemas"]["Poll"];

const client = createClient<paths>();

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

export const createPoll = async (pollData: CreatePollRequest): Promise<Poll> => {
  const { error, data } = await client.POST("/api/polls", {
    body: pollData,
  });
  if (error) {
    raiseApiError(error);
  }
  if (!data) {
    throw new Error("Failed to create poll");
  }
  return data;
};
