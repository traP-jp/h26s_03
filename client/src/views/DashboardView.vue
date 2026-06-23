<template>
  <BasePanel>
    <h1>Dashboard</h1>
    <p class="subtitle">POST /api/initialize でデータ初期化し、全結合クエリ結果を表示します。</p>

    <p v-if="errorMessage" class="notice">{{ errorMessage }}</p>

    <div class="row">
      <BaseButton @click="onInitialize">Initialize</BaseButton>
      <BaseButton variant="secondary" @click="loadFeed">Reload Feed</BaseButton>
    </div>

    <form class="row" @submit.prevent="onSubmitTask">
      <input v-model="taskTitle" class="input" placeholder="New task title" required />
      <select v-model.number="selectedMemberId" class="select" required>
        <option :value="0">Assignee</option>
        <option v-for="m in members" :key="m.id" :value="m.id">
          {{ m.name }}
        </option>
      </select>
      <BaseButton type="submit">Add Task</BaseButton>
    </form>

    <FeedList :items="feed" />
  </BasePanel>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";

import BaseButton from "../components/BaseButton.vue";
import BasePanel from "../components/BasePanel.vue";
import FeedList from "../components/FeedList.vue";
import {
  addTask,
  getFeed,
  getMembers,
  initializeData,
  type FeedItem,
  type Member,
} from "../lib/api";

const feed = ref<FeedItem[]>([]);
const members = ref<Member[]>([]);
const taskTitle = ref("");
const selectedMemberId = ref(0);
const errorMessage = ref("");

async function loadFeed() {
  errorMessage.value = "";
  try {
    feed.value = await getFeed();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

async function loadMembers() {
  errorMessage.value = "";
  try {
    members.value = await getMembers();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

async function onInitialize() {
  errorMessage.value = "";
  try {
    await initializeData();
    await Promise.all([loadFeed(), loadMembers()]);
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

async function onSubmitTask() {
  if (!selectedMemberId.value) {
    errorMessage.value = "member を選択してください";
    return;
  }

  errorMessage.value = "";
  try {
    await addTask(taskTitle.value, selectedMemberId.value);
    taskTitle.value = "";
    await loadFeed();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
}

onMounted(async () => {
  await Promise.all([loadFeed(), loadMembers()]);
});
</script>

<style scoped>
h1 {
  margin: 0;
  font-size: 1.55rem;
}

.subtitle {
  margin: 0;
  color: var(--muted);
  font-size: 0.95rem;
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

.input,
.select {
  min-width: 180px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: #fff;
  padding: 10px;
  font: inherit;
}

@media (max-width: 720px) {
  .row > * {
    width: 100%;
  }
}
</style>
