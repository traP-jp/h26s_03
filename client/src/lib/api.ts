import createClient from "openapi-fetch";

import type { paths, components } from "../gen/api-types";

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

export const getPolls = async (): Promise<Poll[]> => {
  const { data, error } = await client.GET("/api/polls");
  if (error) {
    raiseApiError(error);
  }
  return data.data;
};
