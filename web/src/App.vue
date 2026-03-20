<template>
  <div class="min-h-screen bg-vs-dark">
    <!-- Header -->
    <header class="bg-vs-card border-b border-vs-border sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <!-- Left: Logo + Project Info -->
          <div class="flex items-center space-x-3">
            <div
              class="w-10 h-10 rounded-xl bg-gradient-to-br from-vs-primary to-vs-secondary flex items-center justify-center text-xl font-bold text-white shadow-lg shadow-vs-primary/20"
            >
              V
            </div>
            <div>
              <h1 class="text-lg font-bold text-vs-text">VibeScanner</h1>
              <p v-if="scanResult" class="text-xs text-vs-text-muted">
                {{ scanResult.project?.name }} &middot;
                {{ scanResult.project?.files_scanned }} files &middot;
                {{ formatDate(scanResult.timestamp) }}
              </p>
            </div>
          </div>

          <!-- Center: Report Selector -->
          <div class="flex items-center space-x-2">
            <div class="relative" ref="dropdownRef">
              <button
                @click="showReportDropdown = !showReportDropdown"
                class="flex items-center space-x-2 px-3 py-2 bg-vs-darker border border-vs-border rounded-lg hover:border-vs-primary/50 transition-colors text-sm min-w-[240px]"
              >
                <span class="text-vs-text-muted">📋</span>
                <span class="text-vs-text truncate flex-1 text-left">
                  {{ activeReportLabel }}
                </span>
                <span class="text-vs-text-muted text-xs">{{
                  showReportDropdown ? "▲" : "▼"
                }}</span>
              </button>

              <!-- Dropdown -->
              <div
                v-if="showReportDropdown"
                class="absolute top-full left-0 mt-1 w-[360px] bg-vs-card border border-vs-border rounded-xl shadow-2xl shadow-black/50 overflow-hidden z-50"
              >
                <div class="p-2 border-b border-vs-border">
                  <div
                    class="text-xs text-vs-text-muted px-2 py-1 uppercase tracking-wider font-medium"
                  >
                    Lịch sử quét
                  </div>
                </div>
                <div class="max-h-[320px] overflow-y-auto">
                  <button
                    v-for="report in reports"
                    :key="report.filename"
                    @click="selectReport(report)"
                    class="w-full px-3 py-2.5 text-left hover:bg-vs-primary/10 transition-colors flex items-center space-x-3 group"
                    :class="{
                      'bg-vs-primary/5 border-l-2 border-vs-primary':
                        scanResult?.scan_id === report.scan_id,
                    }"
                  >
                    <div class="flex-1 min-w-0">
                      <div class="text-sm text-vs-text truncate font-medium">
                        {{ report.project_name }}
                      </div>
                      <div
                        class="text-xs text-vs-text-muted flex items-center space-x-2 mt-0.5"
                      >
                        <span>{{ formatDate(report.timestamp) }}</span>
                        <span>&middot;</span>
                        <span
                          :class="
                            report.health_score >= 70
                              ? 'text-vs-success'
                              : report.health_score >= 40
                                ? 'text-vs-warning'
                                : 'text-vs-danger'
                          "
                        >
                          {{ report.health_score }} pts
                        </span>
                        <span>&middot;</span>
                        <span>{{ report.finding_count }} issues</span>
                      </div>
                    </div>
                    <span
                      v-if="scanResult?.scan_id === report.scan_id"
                      class="text-vs-primary text-xs"
                      >active</span
                    >
                  </button>
                  <div
                    v-if="reports.length === 0"
                    class="px-3 py-6 text-center text-vs-text-muted text-sm"
                  >
                    Chưa có lịch sử quét
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Right: Actions -->
          <div class="flex items-center space-x-2">
            <button
              @click="handleRefresh"
              :disabled="loading"
              class="px-3 py-2 bg-vs-primary/10 text-vs-primary rounded-lg hover:bg-vs-primary/20 transition-colors text-sm font-medium flex items-center space-x-1.5"
              :class="{ 'opacity-50 cursor-not-allowed': loading }"
            >
              <span :class="{ 'animate-spin': loading }">🔄</span>
              <span>Làm mới</span>
            </button>
            <div
              class="text-xs text-vs-text-muted bg-vs-darker px-2.5 py-1.5 rounded-full border border-vs-border"
            >
              🔒 Local
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <!-- Loading State -->
      <div
        v-if="loading && !scanResult"
        class="flex items-center justify-center py-20"
      >
        <div class="text-center">
          <div
            class="animate-spin rounded-full h-12 w-12 border-b-2 border-vs-primary mx-auto mb-4"
          ></div>
          <p class="text-vs-text-muted">Đang tải dữ liệu scan...</p>
        </div>
      </div>

      <!-- Error State -->
      <div
        v-else-if="error && !scanResult"
        class="bg-vs-danger/10 border border-vs-danger/30 rounded-xl p-6 text-center"
      >
        <p class="text-vs-danger">{{ error }}</p>
        <button
          @click="loadScanData"
          class="mt-4 px-4 py-2 bg-vs-danger/20 text-vs-danger rounded-lg hover:bg-vs-danger/30 transition-colors"
        >
          Thử lại
        </button>
      </div>

      <!-- Dashboard Content -->
      <div v-else-if="scanResult" class="space-y-5">
        <!-- Health Score Cards -->
        <HealthScore :score="scanResult.health_score" />

        <!-- Summary Bar -->
        <SummaryBar :summary="scanResult.summary" />

        <!-- Filters -->
        <FilterBar :filters="activeFilters" @update:filters="updateFilters" />

        <!-- Findings List -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-semibold text-vs-text">
              Phát hiện ({{ filteredFindings.length }})
            </h2>
            <div class="flex items-center space-x-2">
              <span class="text-sm text-vs-text-muted">Sắp xếp:</span>
              <select
                v-model="sortBy"
                class="bg-vs-card border border-vs-border rounded-lg px-3 py-1.5 text-sm text-vs-text focus:outline-none focus:border-vs-primary"
              >
                <option value="severity">Mức độ nghiêm trọng</option>
                <option value="category">Danh mục</option>
                <option value="file">File</option>
              </select>
            </div>
          </div>

          <div
            v-if="filteredFindings.length === 0"
            class="text-center py-12 bg-vs-card rounded-xl border border-vs-border"
          >
            <p class="text-vs-text-muted text-lg mb-1">
              Không có findings nào khớp
            </p>
            <p class="text-vs-text-muted text-sm">
              Thử thay đổi bộ lọc hoặc từ khóa tìm kiếm
            </p>
          </div>

          <FindingCard
            v-for="finding in filteredFindings"
            :key="finding.id"
            :finding="finding"
            @explain="requestAIExplanation"
          />
        </div>
      </div>
    </main>

    <!-- AI Explanation Modal -->
    <AIModal
      v-if="selectedFinding"
      :finding="selectedFinding"
      @close="selectedFinding = null"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import AIModal from "./components/AIModal.vue";
import FilterBar from "./components/FilterBar.vue";
import FindingCard from "./components/FindingCard.vue";
import HealthScore from "./components/HealthScore.vue";
import SummaryBar from "./components/SummaryBar.vue";
import { useScanStore } from "./stores/scan";

const store = useScanStore();

const loading = computed(() => store.loading);
const error = computed(() => store.error);
const scanResult = computed(() => store.scanResult);
const reports = computed(() => store.reports);
const selectedFinding = ref(null);
const sortBy = ref("severity");
const showReportDropdown = ref(false);
const dropdownRef = ref(null);

const activeFilters = ref({
  severity: [],
  category: [],
  search: "",
});

const activeReportLabel = computed(() => {
  if (!scanResult.value) return "Chọn báo cáo...";
  const name = scanResult.value.project?.name || "Unknown";
  const date = formatDate(scanResult.value.timestamp);
  return `${name} - ${date}`;
});

const loadScanData = () => store.loadScanData();

const handleRefresh = async () => {
  await store.refreshToLatest();
};

const selectReport = async (report) => {
  showReportDropdown.value = false;
  await store.switchReport(report.filename);
};

const updateFilters = (filters) => {
  activeFilters.value = filters;
};

const filteredFindings = computed(() => {
  if (!scanResult.value) return [];
  let findings = [...scanResult.value.findings];

  if (activeFilters.value.severity.length > 0) {
    findings = findings.filter((f) =>
      activeFilters.value.severity.includes(f.severity),
    );
  }
  if (activeFilters.value.category.length > 0) {
    findings = findings.filter((f) =>
      activeFilters.value.category.includes(f.category),
    );
  }
  if (activeFilters.value.search) {
    const search = activeFilters.value.search.toLowerCase();
    findings = findings.filter(
      (f) =>
        f.title?.toLowerCase().includes(search) ||
        f.message?.toLowerCase().includes(search) ||
        f.file?.toLowerCase().includes(search),
    );
  }

  const severityOrder = { critical: 0, high: 1, medium: 2, low: 3, info: 4 };
  switch (sortBy.value) {
    case "severity":
      findings.sort(
        (a, b) =>
          (severityOrder[a.severity] ?? 5) - (severityOrder[b.severity] ?? 5),
      );
      break;
    case "category":
      findings.sort((a, b) =>
        (a.category || "").localeCompare(b.category || ""),
      );
      break;
    case "file":
      findings.sort((a, b) => (a.file || "").localeCompare(b.file || ""));
      break;
  }
  return findings;
});

const requestAIExplanation = (finding) => {
  selectedFinding.value = finding;
};

const formatDate = (timestamp) => {
  if (!timestamp) return "";
  return new Date(timestamp).toLocaleString("vi-VN", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

// Close dropdown on outside click
const handleClickOutside = (e) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    showReportDropdown.value = false;
  }
};

onMounted(async () => {
  document.addEventListener("click", handleClickOutside);
  await store.loadScanData();
  await store.loadReports();
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>
