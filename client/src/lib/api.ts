import createClient from "openapi-fetch";

import type { paths, components } from "../gen/api-types";

type CreatePollRequest = components["schemas"]["CreatePollRequest"];
export type Poll = components["schemas"]["Poll"];

const client = createClient<paths>();

function raiseApiError(error: unknown): never {
  if (error instanceof Error) {
    throw error;
  }
  throw new Error("request failed");
}

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

export const getPolls = async (): Promise<Poll[]> => {
  const { data, error } = await client.GET("/api/polls");
  if (error) {
    raiseApiError(error);
  }
  return data.data;
};

export const getPoll = async (id: number): Promise<Poll> => {
  const { data, error } = await client.GET("/api/polls/{id}", { params: { path: { id } } });
  if (error) {
    raiseApiError(error);
  }
  if (data === undefined) {
    raiseApiError(new Error(`Poll with id=${id} not found`));
  }
  return data;
};
export const updatePoll = async (id: number, result: number): Promise<Poll> => {
  const { data, error } = await client.PATCH("/api/polls/{id}", {
    params: {
      path: { id },
    },
    body: {
      result,
    },
  });

  if (error) {
    raiseApiError(error);
  }

  if (data === undefined) {
    raiseApiError(new Error(`Poll with id=${id} not found`));
  }

  return data;
};
