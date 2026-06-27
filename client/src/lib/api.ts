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

export const getPolls = async (): Promise<Poll[]> => {
  const { data, error } = await client.GET("/api/polls");
  if (error) {
    raiseApiError(error);
  }
  return data.data;
};

export const getPoll = async (pollId: number) => {
  const { data, error } = await client.GET("/api/polls/{id}", {
    params: {
      path: {
        id: pollId,
      },
    },
  });

  if (error) {
    raiseApiError(error);
  }

  if (!data) {
    throw new Error("poll data is empty");
  }

  return data;
};

export const getVotes = async (pollId: number) => {
  const { data, error } = await client.GET("/api/polls/{id}/votes", {
    params: {
      path: {
        id: pollId,
      },
    },
  });

  if (error) {
    raiseApiError(error);
  }

  return data?.data ?? [];
};

export const createVote = async (pollId: number, choice: number, bet: number) => {
  const { data, error } = await client.POST("/api/polls/{id}/votes", {
    params: {
      path: {
        id: pollId,
      },
    },
    body: {
      choice,
      bet,
    },
  });

  if (error) {
    raiseApiError(error);
  }

  return data;
};

export const deleteVote = async (poll_id: number, vote_id: number) => {
  const { error } = await client.DELETE("/api/polls/{poll_id}/votes/{vote_id}", {
    params: {
      path: {
        poll_id,
        vote_id,
      },
    },
  });

  if (error) {
    raiseApiError(error);
  }
};

export const getMe = async () => {
  const { data, error } = await client.GET("/api/me");
  if (error) {
    raiseApiError(error);
  }

  return data;
};
