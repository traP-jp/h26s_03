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

export const getMe = async () => {
  const { data, error } = await client.GET("/api/me");
  if (error) {
    raiseApiError(error);
  }
  if (data === undefined) {
    raiseApiError(new Error(`Me not found`));
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

export const deleteVote = async (pollId: number, voteId: number) => {
  const { error } = await client.DELETE("/api/polls/{poll_id}/votes/{vote_id}", {
    params: {
      path: {
        poll_id: pollId,
        vote_id: voteId,
      },
    },
  });

  if (error) {
    raiseApiError(error);
  }
};
