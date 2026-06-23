<template>
  <BasePanel>
    <h1>Members</h1>
    <p class="subtitle">バックエンドのサンプルメンバー一覧です。</p>

    <p v-if="errorMessage" class="notice">{{ errorMessage }}</p>

    <ul class="list">
      <li v-for="member in members" :key="member.id" class="card">
        <p class="card-title">{{ member.name }}</p>
        <p class="meta">id: {{ member.id }}</p>
      </li>
    </ul>
  </BasePanel>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";

import BasePanel from "../components/BasePanel.vue";
import { getMembers, type Member } from "../lib/api";

const members = ref<Member[]>([]);
const errorMessage = ref("");

onMounted(async () => {
  try {
    members.value = await getMembers();
  } catch (err) {
    errorMessage.value = (err as Error).message;
  }
});
</script>

<style scoped>
.subtitle {
  margin: 0;
  color: var(--muted);
  font-size: 0.95rem;
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

.list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.card {
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px;
  background: linear-gradient(130deg, #fff 0%, #f9fdf6 100%);
}

.card-title {
  margin: 0 0 4px;
  font-weight: 700;
}

.meta {
  margin: 0;
  font-size: 0.86rem;
  color: var(--muted);
}
</style>
