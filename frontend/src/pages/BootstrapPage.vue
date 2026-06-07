<template>
  <main class="bootstrap-page">
    <AppStatusCard :state="view.state" :title="view.title" :message="view.message" />
  </main>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import AppStatusCard from '@/components/AppStatusCard.vue'
import { loadBootstrapStatus, useBootstrapStore } from '@/stores/bootstrapStore'

const view = useBootstrapStore()

onMounted(() => {
  void loadBootstrapStatus()
})
</script>

<style scoped>
.bootstrap-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 2rem;
  background: linear-gradient(135deg, #f7f9fc, #edf4ff);
}
</style>
