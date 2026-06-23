<template>
  <section class="panel stack">
    <slot />
  </section>
</template>

<style scoped>
.panel {
  background: var(--bg-card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 20px;
}

.stack {
  display: grid;
  gap: 14px;
}
</style>
