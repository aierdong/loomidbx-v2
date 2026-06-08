<template>
  <section class="status-card" :data-state="state">
    <p class="eyebrow">{{ state }}</p>
    <h2>{{ title }}</h2>
    <p>{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  state: "loading" | "ready" | "error";
  title: string;
  message: string;
}>();
</script>

<style scoped>
.status-card {
  max-width: 36rem;
  padding: 1.5rem;
  border: 1px solid #d9e2ef;
  border-radius: 1rem;
  background: #ffffff;
  box-shadow: 0 1rem 2rem rgb(23 32 51 / 8%);
}

.eyebrow {
  margin: 0 0 0.5rem;
  color: #476582;
  font-size: 0.75rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
</style>
