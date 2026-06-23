<template>
  <ul class="list">
    <li v-for="item in items" :key="item.task_id" class="card">
      <p class="card-title">{{ item.task_title }}</p>
      <p class="meta">
        Project: {{ item.project_name }} / Owner: {{ item.owner_name }} /
        <span class="badge">{{ item.task_status }}</span>
      </p>
    </li>
  </ul>
</template>

<script setup lang="ts">
import type { FeedItem } from "../lib/api";

defineProps<{
  items: FeedItem[];
}>();
</script>

<style scoped>
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

.badge {
  display: inline-flex;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid var(--line);
  font-size: 0.75rem;
  background: #fff;
}
</style>
