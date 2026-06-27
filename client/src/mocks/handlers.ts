import { HttpResponse } from "msw";
import { createOpenApiHttp } from "openapi-msw";
import * as v from "valibot";

import type { components, paths } from "../gen/api-types";

type Poll = components["schemas"]["Poll"];
type CreatePollRequest = components["schemas"]["CreatePollRequest"];
type UpdatePollRequest = components["schemas"]["UpdatePollRequest"];
type PollsResponse = components["schemas"]["PollsResponse"];
type Vote = components["schemas"]["Vote"];
type CreateVoteRequest = components["schemas"]["CreateVoteRequest"];
type VotesResponse = components["schemas"]["VotesResponse"];
type Me = components["schemas"]["Me"];

type MockVote = Vote & { poll_id: number };

const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";
const http = createOpenApiHttp<paths>({ baseUrl: apiBase });

const choiceSchema = v.union([v.literal(1), v.literal(2)]);
const createPollRequestSchema: v.GenericSchema<CreatePollRequest> = v.object({
  name: v.pipe(v.string(), v.nonEmpty()),
  choice1: v.pipe(v.string(), v.nonEmpty()),
  choice2: v.pipe(v.string(), v.nonEmpty()),
  due: v.optional(v.nullable(v.string())),
});
const updatePollRequestSchema: v.GenericSchema<UpdatePollRequest> = v.object({
  name: v.optional(v.string()),
  choice1: v.optional(v.string()),
  choice2: v.optional(v.string()),
  result: v.optional(v.nullable(choiceSchema)),
  due: v.optional(v.nullable(v.string())),
});
const createVoteRequestSchema: v.GenericSchema<CreateVoteRequest> = v.object({
  choice: choiceSchema,
  bet: v.pipe(v.number(), v.integer()),
});

const initialPolls: Poll[] = [
  {
    id: 1,
    name: "決勝で勝つのはどっち？",
    choice1: "チームA",
    choice2: "チームB",
    result: null,
    due: null,
    created_by: "developer",
    created_at: "2026-06-27T00:00:00.000Z",
  },
];

const initialVotes: MockVote[] = [];

const createInitialPolls = (): Poll[] => {
  return initialPolls.map((poll) => ({ ...poll }));
};

const createInitialVotes = (): MockVote[] => {
  return initialVotes.map((vote) => ({ ...vote }));
};

const getNextPollId = (currentPolls: Poll[]): number => {
  const maxId = currentPolls.reduce((max, poll) => Math.max(max, poll.id), 0);
  return maxId + 1;
};

const getNextVoteId = (currentVotes: MockVote[]): number => {
  const maxId = currentVotes.reduce((max, vote) => Math.max(max, vote.id), 0);
  return maxId + 1;
};

let polls = createInitialPolls();
let votes = createInitialVotes();
let nextPollId = getNextPollId(polls);
let nextVoteId = getNextVoteId(votes);

const resetPolls = (): void => {
  polls = createInitialPolls();
  votes = createInitialVotes();
  nextPollId = getNextPollId(polls);
  nextVoteId = getNextVoteId(votes);
};

const getUsername = (request: Request): string => {
  return request.headers.get("x-forwarded-user") ?? "developer";
};

const jsonError = (message: string, status: number) => {
  return HttpResponse.json({ message }, { status });
};

const parseJsonBody = async <TSchema extends v.GenericSchema>(
  request: Request,
  schema: TSchema,
) => {
  const body = await request.json().catch(() => undefined);
  return v.safeParse(schema, body);
};

export const handlers = [
  http.post("/api/initialize", ({ response }) => {
    resetPolls();
    return response(204).empty();
  }),

  http.get("/api/polls", ({ response }) => {
    const body: PollsResponse = {
      data: polls.map((poll) => ({ ...poll })),
    };

    return response(200).json(body);
  }),

  http.post("/api/polls", async ({ request, response }) => {
    const parsed = await parseJsonBody(request, createPollRequestSchema);
    if (!parsed.success) {
      return response.untyped(jsonError("name, choice1 and choice2 are required", 400));
    }
    const body = parsed.output;

    const poll: Poll = {
      id: nextPollId,
      name: body.name,
      choice1: body.choice1,
      choice2: body.choice2,
      result: null,
      due: body.due ?? null,
      created_by: getUsername(request),
      created_at: new Date().toISOString(),
    };

    polls.push(poll);
    nextPollId += 1;

    return response(201).json({ ...poll });
  }),

  http.get("/api/polls/{id}", ({ params, response }) => {
    const poll = polls.find((item) => item.id === Number(params.id));
    if (!poll) {
      return response.untyped(jsonError("poll not found", 404));
    }

    return response(200).json({ ...poll });
  }),

  http.patch("/api/polls/{id}", async ({ params, request, response }) => {
    const poll = polls.find((item) => item.id === Number(params.id));
    if (!poll) {
      return response.untyped(jsonError("poll not found", 404));
    }
    if (poll.created_by !== getUsername(request)) {
      return response.untyped(jsonError("poll owner mismatch", 403));
    }

    const parsed = await parseJsonBody(request, updatePollRequestSchema);
    if (!parsed.success) {
      return response.untyped(jsonError("request body is required", 400));
    }
    const body = parsed.output;

    if (body.name !== undefined) poll.name = body.name;
    if (body.choice1 !== undefined) poll.choice1 = body.choice1;
    if (body.choice2 !== undefined) poll.choice2 = body.choice2;
    if (body.result !== undefined) poll.result = body.result;
    if (body.due !== undefined) poll.due = body.due;

    return response(200).json({ ...poll });
  }),

  http.delete("/api/polls/{id}", ({ params, request, response }) => {
    const pollIndex = polls.findIndex((item) => item.id === Number(params.id));
    if (pollIndex < 0) {
      return response.untyped(jsonError("poll not found", 404));
    }
    const poll = polls[pollIndex];
    if (poll.created_by !== getUsername(request)) {
      return response.untyped(jsonError("poll owner mismatch", 403));
    }

    polls.splice(pollIndex, 1);
    votes = votes.filter((vote) => vote.poll_id !== poll.id);
    return response(204).empty();
  }),

  http.get("/api/polls/{id}/votes", ({ params, response }) => {
    const pollID = Number(params.id);
    if (!polls.some((poll) => poll.id === pollID)) {
      return response.untyped(jsonError("poll not found", 404));
    }

    const body: VotesResponse = {
      data: votes
        .filter((vote) => vote.poll_id === pollID)
        .map((vote) => ({
          id: vote.id,
          username: vote.username,
          choice: vote.choice,
          bet: vote.bet,
          created_at: vote.created_at,
        })),
    };

    return response(200).json(body);
  }),

  http.post("/api/polls/{id}/votes", async ({ params, request, response }) => {
    const pollID = Number(params.id);
    if (!polls.some((poll) => poll.id === pollID)) {
      return response.untyped(jsonError("poll not found", 404));
    }

    const parsed = await parseJsonBody(request, createVoteRequestSchema);
    if (!parsed.success) {
      return response.untyped(jsonError("choice must be 1 or 2, and bet is required", 400));
    }
    const body = parsed.output;

    const username = getUsername(request);
    if (votes.some((vote) => vote.poll_id === pollID && vote.username === username)) {
      return response.untyped(jsonError("already voted", 409));
    }

    const vote: MockVote = {
      id: nextVoteId,
      poll_id: pollID,
      username,
      choice: body.choice,
      bet: body.bet,
      created_at: new Date().toISOString(),
    };
    votes.push(vote);
    nextVoteId += 1;

    return response(201).json({
      id: vote.id,
      username: vote.username,
      choice: vote.choice,
      bet: vote.bet,
      created_at: vote.created_at,
    });
  }),

  http.delete("/api/polls/{poll_id}/votes/{vote_id}", ({ params, request, response }) => {
    const pollID = Number(params.poll_id);
    const voteID = Number(params.vote_id);
    const voteIndex = votes.findIndex((vote) => vote.poll_id === pollID && vote.id === voteID);

    if (voteIndex < 0) {
      return response.untyped(jsonError("vote not found", 404));
    }
    if (votes[voteIndex].username !== getUsername(request)) {
      return response.untyped(jsonError("vote owner mismatch", 403));
    }

    votes.splice(voteIndex, 1);
    return response(204).empty();
  }),

  http.get("/api/me", ({ request, response }) => {
    const body: Me = {
      username: getUsername(request),
      balance: 1000,
    };

    return response(200).json(body);
  }),
];
