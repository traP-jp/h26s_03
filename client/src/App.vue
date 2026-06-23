<template>
  <main class="shell">
    <header class="topbar">
      <div class="brand-block">
        <img class="brand-icon" :src="typescriptLogo" alt="" aria-hidden="true" />
        <p class="brand">Spring Hackathon Template</p>
        <p class="subtitle">Vue + Router + Echo API</p>
      </div>
      <nav class="nav">
        <RouterLink to="/">Dashboard</RouterLink>
        <RouterLink to="/members">Members</RouterLink>
      </nav>
    </header>

    <RouterView />
  </main>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router";

import typescriptLogo from "./assets/typescript.svg";
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

@media (max-width: 720px) {
  .shell {
    padding: 14px 12px 30px;
  }

  .topbar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
