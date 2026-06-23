<template>
  <main class="shell">
    <header class="topbar">
      <div class="brand-block">
        <img class="brand-icon" :src="typescriptLogo" alt="" aria-hidden="true" />
        <p class="brand">Spring Hackathon Template</p>
        <p class="app-subtitle">Vue + Router + Echo API</p>
      </div>
      <nav class="nav">
        <RouterLink to="/">Back</RouterLink>
      </nav>
    </header>

    <BasePanel>
      <h1>Todo</h1>
      <p class="subtitle">POST /api/initialize でデータ初期化し、tasks を表示します。</p>

      <p v-if="errorMessage" class="notice">{{ errorMessage }}</p>

      <div class="row">
        <BaseButton @click="onInitialize">Initialize</BaseButton>
        <BaseButton variant="secondary" @click="loadTasks">Reload Tasks</BaseButton>
      </div>

      <form class="row" @submit.prevent="onSubmitTask">
        <input v-model="taskTitle" class="input" placeholder="New task title" required />
        <BaseButton type="submit">Add Task</BaseButton>
      </form>

      <TaskList :tasks="tasks" />
    </BasePanel>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";

import typescriptLogo from "../assets/typescript.svg";
import BaseButton from "../components/BaseButton.vue";
import BasePanel from "../components/BasePanel.vue";
import TaskList from "../components/TaskList.vue";
import { addTask, getTasks, initializeData, type Task } from "../lib/api";

const tasks = ref<Task[]>([]);
const taskTitle = ref("");
const errorMessage = ref("");

async function loadTasks() {
  errorMessage.value = "";
  try {
    tasks.value = await getTasks();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

async function onInitialize() {
  errorMessage.value = "";
  try {
    await initializeData();
    await loadTasks();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

async function onSubmitTask() {
  errorMessage.value = "";
  try {
    await addTask(taskTitle.value);
    taskTitle.value = "";
    await loadTasks();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

onMounted(async () => {
  await loadTasks();
});
</script>

<style scoped>
.shell {
  width: min(980px, 100%);
  margin: 0 auto;
  padding: 24px 16px 40px;
}

.topbar {
  display: flex;
  gap: 16px;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 18px;
}

.brand-block {
  display: grid;
  grid-template-columns: 34px 1fr;
  gap: 2px 10px;
  align-items: center;
}

.brand-icon {
  grid-row: span 2;
  width: 34px;
  height: 34px;
}

.brand {
  margin: 0;
  font-family: "M PLUS Rounded 1c", "Noto Sans JP", sans-serif;
  font-weight: 700;
  font-size: 1.1rem;
}

.app-subtitle,
.subtitle {
  margin: 0;
  color: var(--muted);
  font-size: 0.95rem;
}

.nav {
  display: flex;
  gap: 8px;
}

.nav a {
  text-decoration: none;
  border: 1px solid var(--line);
  background: var(--bg-card);
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 0.9rem;
}

.nav a.router-link-active {
  border-color: var(--primary);
  color: var(--primary-strong);
  background: var(--bg-accent);
}

h1 {
  margin: 0;
  font-size: 1.55rem;
}

.notice {
  margin: 0;
  padding: 10px;
  border-radius: 10px;
  border: 1px dashed var(--line);
  background: #fcfff9;
  font-size: 0.9rem;
}

.row {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.input {
  min-width: 180px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: #fff;
  padding: 10px;
  font: inherit;
}

@media (max-width: 720px) {
  .shell {
    padding: 14px 12px 30px;
  }

  .topbar {
    flex-direction: column;
    align-items: flex-start;
  }

  .row > * {
    width: 100%;
  }
}
</style>
