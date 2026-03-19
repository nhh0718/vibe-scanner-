<template>
  <div class="flex flex-wrap gap-3">
    <!-- Critical -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2">
      <span class="text-2xl font-bold text-vs-danger">{{ summary.critical }}</span>
      <span class="text-sm text-vs-text-muted">Critical</span>
    </div>

    <!-- High -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2">
      <span class="text-2xl font-bold text-orange-500">{{ summary.high }}</span>
      <span class="text-sm text-vs-text-muted">High</span>
    </div>

    <!-- Medium -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2">
      <span class="text-2xl font-bold text-vs-warning">{{ summary.medium }}</span>
      <span class="text-sm text-vs-text-muted">Medium</span>
    </div>

    <!-- Low -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2">
      <span class="text-2xl font-bold text-vs-info">{{ summary.low }}</span>
      <span class="text-sm text-vs-text-muted">Low</span>
    </div>

    <!-- Info -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2">
      <span class="text-2xl font-bold text-gray-500">{{ summary.info }}</span>
      <span class="text-sm text-vs-text-muted">Info</span>
    </div>

    <!-- Total -->
    <div class="bg-vs-card border border-vs-border rounded-lg px-4 py-3 flex items-center space-x-2 ml-auto">
      <span class="text-2xl font-bold text-vs-text">{{ summary.total }}</span>
      <span class="text-sm text-vs-text-muted">Tổng</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  summary: {
    type: Object,
    required: true,
    default: () => ({
      critical: 0,
      high: 0,
      medium: 0,
      low: 0,
      info: 0,
      total: 0
    })
  }
})
</script>
