import { HttpResponse } from "msw";
import { createOpenApiHttp } from "openapi-msw";

import type { components, paths } from "../gen/api-types";

type Task = components["schemas"]["Task"];
type CreateTaskRequest = components["schemas"]["CreateTaskRequest"];
type TasksResponse = components["schemas"]["TasksResponse"];

const apiBase = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";
const http = createOpenApiHttp<paths>({ baseUrl: apiBase });

const initialTasks: Task[] = [
  { id: 1, title: "トップページの構成を考える", status: "todo" },
  { id: 2, title: "タスク追加フォームを作る", status: "doing" },
  { id: 3, title: "初期化APIをつなぐ", status: "done" },
];

let tasks = createInitialTasks();
let nextTaskId = getNextTaskId(tasks);

function createInitialTasks(): Task[] {
  return initialTasks.map((task) => ({ ...task }));
}

function getNextTaskId(currentTasks: Task[]): number {
  const maxId = currentTasks.reduce((max, task) => Math.max(max, task.id), 0);
  return maxId + 1;
}

function resetTasks() {
  tasks = createInitialTasks();
  nextTaskId = getNextTaskId(tasks);
}

export const handlers = [
  http.post("/api/initialize", ({ response }) => {
    resetTasks();
    return response(204).empty();
  }),

  http.get("/api/tasks", ({ response }) => {
    const body: TasksResponse = {
      data: tasks.map((task) => ({ ...task })),
    };

    return response(200).json(body);
  }),

  http.post("/api/tasks", async ({ request, response }) => {
    const body = (await request.json().catch(() => null)) as Partial<CreateTaskRequest> | null;
    const title = body?.title;

    if (typeof title !== "string" || title === "") {
      return response.untyped(HttpResponse.json({ message: "title is required" }, { status: 400 }));
    }

    tasks.push({
      id: nextTaskId,
      title,
      status: "todo",
    });
    nextTaskId += 1;

    return response(201).empty();
  }),
];
