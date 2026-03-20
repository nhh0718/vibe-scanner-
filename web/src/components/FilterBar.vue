<template>
  <div class="filter-bar">
    <div class="filter-pills">
      <button
        class="fpill"
        :class="{ active: localFilters.severity.length === 0 }"
        @click="clearFilters"
      >
        Tất cả <span class="fpill-n">{{ summary.total || 0 }}</span>
      </button>
      <button
        class="fpill"
        :class="{ active: localFilters.severity.includes('critical') }"
        @click="toggleSeverity('critical')"
      >
        <span class="fpill-dot" style="background: var(--red)"></span>
        Critical <span class="fpill-n">{{ summary.critical || 0 }}</span>
      </button>
      <button
        class="fpill"
        :class="{ active: localFilters.severity.includes('high') }"
        @click="toggleSeverity('high')"
      >
        <span class="fpill-dot" style="background: var(--ora)"></span>
        High <span class="fpill-n">{{ summary.high || 0 }}</span>
      </button>
      <button
        class="fpill"
        :class="{ active: localFilters.severity.includes('medium') }"
        @click="toggleSeverity('medium')"
      >
        <span class="fpill-dot" style="background: var(--yel)"></span>
        Medium <span class="fpill-n">{{ summary.medium || 0 }}</span>
      </button>
      <button
        class="fpill"
        :class="{ active: localFilters.severity.includes('low') }"
        @click="toggleSeverity('low')"
      >
        <span class="fpill-dot" style="background: var(--blu)"></span>
        Low <span class="fpill-n">{{ summary.low || 0 }}</span>
      </button>
    </div>

    <div class="filter-search">
      <div class="filter-search-icon">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="11" cy="11" r="8" />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M21 21l-4.35-4.35"
          />
        </svg>
      </div>
      <input
        class="search-input"
        type="text"
        v-model="localFilters.search"
        @input="updateFilters"
        placeholder="Tìm file, rule ID, mô tả..."
      />
    </div>

    <div class="filter-right">
      <button class="btn">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
          width="13"
          height="13"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12"
          />
        </svg>
        Sắp xếp
      </button>
      <button class="btn">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
          width="13"
          height="13"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z"
          />
        </svg>
        Lọc
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  filters: {
    type: Object,
    default: () => ({
      severity: [],
      category: [],
      search: "",
    }),
  },
  summary: {
    type: Object,
    default: () => ({
      critical: 0,
      high: 0,
      medium: 0,
      low: 0,
      total: 0,
    }),
  },
});

const emit = defineEmits(["update:filters"]);

const localFilters = ref({
  severity: [...props.filters.severity],
  category: [...props.filters.category],
  search: props.filters.search,
});

watch(
  () => props.filters,
  (newFilters) => {
    localFilters.value = {
      severity: [...newFilters.severity],
      category: [...newFilters.category],
      search: newFilters.search,
    };
  },
  { deep: true },
);

const toggleSeverity = (sev) => {
  const idx = localFilters.value.severity.indexOf(sev);
  if (idx > -1) {
    localFilters.value.severity.splice(idx, 1);
  } else {
    localFilters.value.severity.push(sev);
  }
  updateFilters();
};

const clearFilters = () => {
  localFilters.value = {
    severity: [],
    category: localFilters.value.category,
    search: localFilters.value.search,
  };
  updateFilters();
};

const updateFilters = () => {
  emit("update:filters", { ...localFilters.value });
};
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 24px 12px;
}

.filter-pills {
  display: flex;
  background: var(--bg2);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
  padding: 3px;
  gap: 2px;
}

.filter-search {
  flex: 1;
  max-width: 260px;
  position: relative;
}

.filter-search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text3);
  pointer-events: none;
}
.filter-search-icon svg {
  width: 13px;
  height: 13px;
}

.search-input {
  width: 100%;
  background: var(--bg2);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
  color: var(--text1);
  font-family: var(--sans);
  font-size: 12px;
  padding: 7px 10px 7px 32px;
  outline: none;
  transition: border-color 0.15s;
}
.search-input::placeholder {
  color: var(--text3);
}
.search-input:focus {
  border-color: var(--acc);
  background: var(--bg2);
}

.filter-right {
  margin-left: auto;
  display: flex;
  gap: 6px;
}
</style>
