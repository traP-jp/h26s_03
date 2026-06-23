<template>
  <button :class="classes" class="button" :type="type" @click="$emit('click')">
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    type?: "button" | "submit";
    variant?: "primary" | "secondary" | "danger";
  }>(),
  {
    type: "button",
    variant: "primary",
  },
);

defineEmits<{
  click: [];
}>();

const classes = computed(() => {
  return {
    secondary: props.variant === "secondary",
    danger: props.variant === "danger",
  };
});
</script>

<style scoped>
.button {
  border-radius: 10px;
  border: 1px solid var(--primary);
  background: var(--primary);
  color: #fff;
  padding: 10px 14px;
  font: inherit;
  cursor: pointer;
  transition:
    background 0.12s ease;
}

.button:hover {
  background: var(--primary-strong);
}

.button.secondary {
  border-color: var(--line);
  background: #fff;
  color: var(--text);
}

.button.danger {
  border-color: var(--danger);
  background: var(--danger);
}
</style>
