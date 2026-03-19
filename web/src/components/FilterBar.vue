<template>
  <div class="bg-vs-card border border-vs-border rounded-xl p-4 space-y-4">
    <!-- Search -->
    <div class="flex items-center space-x-2">
      <span class="text-vs-text-muted">🔍</span>
      <input
        v-model="localFilters.search"
        type="text"
        placeholder="Tìm kiếm findings..."
        class="flex-1 bg-vs-darker border border-vs-border rounded-lg px-4 py-2 text-vs-text placeholder-vs-text-muted focus:outline-none focus:border-vs-primary"
        @input="updateFilters"
      />
    </div>

    <div class="flex flex-wrap gap-4">
      <!-- Severity Filter -->
      <div class="space-y-2">
        <label class="text-sm text-vs-text-muted font-medium">Severity</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="sev in severities"
            :key="sev.value"
            @click="toggleSeverity(sev.value)"
            class="px-3 py-1.5 rounded-lg text-sm transition-colors border"
            :class="localFilters.severity.includes(sev.value)
              ? sev.activeClass
              : 'bg-vs-darker border-vs-border text-vs-text-muted'"
          >
            {{ sev.label }}
          </button>
        </div>
      </div>

      <!-- Category Filter -->
      <div class="space-y-2">
        <label class="text-sm text-vs-text-muted font-medium">Danh mục</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="cat in categories"
            :key="cat.value"
            @click="toggleCategory(cat.value)"
            class="px-3 py-1.5 rounded-lg text-sm transition-colors border"
            :class="localFilters.category.includes(cat.value)
              ? cat.activeClass
              : 'bg-vs-darker border-vs-border text-vs-text-muted'"
          >
            {{ cat.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Clear Filters -->
    <div v-if="hasActiveFilters" class="flex justify-end">
      <button
        @click="clearFilters"
        class="text-sm text-vs-text-muted hover:text-vs-primary transition-colors"
      >
        ✕ Xóa bộ lọc
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  filters: {
    type: Object,
    default: () => ({
      severity: [],
      category: [],
      search: ''
    })
  }
})

const emit = defineEmits(['update:filters'])

const localFilters = ref({
  severity: [...props.filters.severity],
  category: [...props.filters.category],
  search: props.filters.search
})

watch(() => props.filters, (newFilters) => {
  localFilters.value = {
    severity: [...newFilters.severity],
    category: [...newFilters.category],
    search: newFilters.search
  }
}, { deep: true })

const severities = [
  { value: 'critical', label: 'Critical', activeClass: 'bg-vs-danger/20 border-vs-danger text-vs-danger' },
  { value: 'high', label: 'High', activeClass: 'bg-orange-500/20 border-orange-500 text-orange-500' },
  { value: 'medium', label: 'Medium', activeClass: 'bg-vs-warning/20 border-vs-warning text-vs-warning' },
  { value: 'low', label: 'Low', activeClass: 'bg-vs-info/20 border-vs-info text-vs-info' },
  { value: 'info', label: 'Info', activeClass: 'bg-gray-500/20 border-gray-500 text-gray-500' }
]

const categories = [
  { value: 'security', label: '🔐 Bảo mật', activeClass: 'bg-vs-danger/20 border-vs-danger text-vs-danger' },
  { value: 'quality', label: '✨ Chất lượng', activeClass: 'bg-vs-info/20 border-vs-info text-vs-info' },
  { value: 'architecture', label: '🏗️ Kiến trúc', activeClass: 'bg-vs-secondary/20 border-vs-secondary text-vs-secondary' },
  { value: 'secrets', label: '🔑 Secrets', activeClass: 'bg-pink-500/20 border-pink-500 text-pink-500' },
  { value: 'performance', label: '⚡ Hiệu năng', activeClass: 'bg-vs-success/20 border-vs-success text-vs-success' }
]

const hasActiveFilters = computed(() => {
  return localFilters.value.severity.length > 0 ||
         localFilters.value.category.length > 0 ||
         localFilters.value.search.length > 0
})

const toggleSeverity = (sev) => {
  const idx = localFilters.value.severity.indexOf(sev)
  if (idx > -1) {
    localFilters.value.severity.splice(idx, 1)
  } else {
    localFilters.value.severity.push(sev)
  }
  updateFilters()
}

const toggleCategory = (cat) => {
  const idx = localFilters.value.category.indexOf(cat)
  if (idx > -1) {
    localFilters.value.category.splice(idx, 1)
  } else {
    localFilters.value.category.push(cat)
  }
  updateFilters()
}

const clearFilters = () => {
  localFilters.value = {
    severity: [],
    category: [],
    search: ''
  }
  updateFilters()
}

const updateFilters = () => {
  emit('update:filters', { ...localFilters.value })
}
</script>
