<template>
  <div class="app">
    <!-- ════════════════════════════════
         TOPBAR
    ════════════════════════════════ -->
    <header class="topbar">
      <div class="tb-logo">
        <div class="logo-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 12h6m-3-3v6M12 3C7.03 3 3 7.03 3 12s4.03 9 9 9 9-4.03 9-9-4.03-9-9-9z"
            />
          </svg>
        </div>
        <div>
          <div class="logo-name">VibeScanner</div>
          <div class="logo-tag">Code Doctor</div>
        </div>
      </div>

      <div class="tb-divider"></div>

      <div class="tb-project" v-if="scanResult">
        <div class="tb-project-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            width="12"
            height="12"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M3 7a2 2 0 012-2h14a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"
            />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M3 7l9 6 9-6"
            />
          </svg>
        </div>
        <div>
          <div class="tb-proj-name">
            {{ scanResult.project?.name || "my-project" }}
          </div>
          <div class="tb-proj-sub">
            scan #{{ activeReportId?.substring(0, 6) || "latest" }} &middot;
            {{ formatDate(scanResult.timestamp) }}
          </div>
        </div>
      </div>

      <div class="tb-status" v-if="!loading && scanResult">
        <div class="tb-status-dot"></div>
        <div class="tb-status-text">
          Hoàn thành &middot; {{ scanResult.project?.files_scanned }} files
        </div>
      </div>
      <div
        class="tb-status"
        style="background: var(--ora-bg); border-color: var(--ora-bg)"
        v-else-if="loading"
      >
        <div
          class="tb-status-dot"
          style="background: var(--ora); box-shadow: 0 0 6px var(--ora)"
        ></div>
        <div class="tb-status-text" style="color: var(--ora)">Đang quét...</div>
      </div>

      <div class="tb-actions">
        <button class="btn" @click="toggleTheme" title="Toggle Theme">
          <svg
            v-if="isDark"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>
          <svg
            v-else
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
            />
          </svg>
        </button>
        <button class="btn" @click="handleRefresh" :disabled="loading">
          <svg
            :class="{ 'animate-spin': loading }"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4 4v5h5M20 20v-5h-5M4 9a9 9 0 0115.457-5.457M20 15a9 9 0 01-15.457 5.457"
            />
          </svg>
          Quét lại
        </button>
      </div>
    </header>

    <!-- ════════════════════════════════
         SIDEBAR
    ════════════════════════════════ -->
    <aside class="sidebar">
      <div style="padding: 12px 12px 4px" v-if="scanResult">
        <HealthScore :score="scanResult.health_score" />
      </div>

      <div class="sb-section">
        <div class="sb-label">Điều hướng</div>
        <div class="sb-item active">
          <div class="sb-icon">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M9 12h6m-3-3v6M12 3C7.03 3 3 7.03 3 12s4.03 9 9 9 9-4.03 9-9-4.03-9-9-9z"
              />
            </svg>
          </div>
          Hồ sơ bệnh án
        </div>
      </div>

      <div class="sb-divider"></div>

      <div class="sb-section" v-if="scanResult">
        <div class="sb-label">Phân loại</div>
        <div
          class="sb-item"
          @click="toggleFilter('category', 'security')"
          :class="{ active: activeFilters.category.includes('security') }"
        >
          <div class="sb-icon" style="color: var(--red)">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
              />
            </svg>
          </div>
          Bảo mật
          <span
            class="sb-badge"
            style="background: var(--red-bg); color: var(--red)"
          >
            {{ getCategoryCount("security") }}
          </span>
        </div>
        <div
          class="sb-item"
          @click="toggleFilter('category', 'quality')"
          :class="{ active: activeFilters.category.includes('quality') }"
        >
          <div class="sb-icon" style="color: var(--ora)">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
              />
            </svg>
          </div>
          Chất lượng
          <span
            class="sb-badge"
            style="background: var(--ora-bg); color: var(--ora)"
          >
            {{ getCategoryCount("quality") }}
          </span>
        </div>
        <div
          class="sb-item"
          @click="toggleFilter('category', 'architecture')"
          :class="{ active: activeFilters.category.includes('architecture') }"
        >
          <div class="sb-icon" style="color: var(--yel)">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <rect x="3" y="3" width="7" height="7" rx="1" />
              <rect x="14" y="3" width="7" height="7" rx="1" />
              <rect x="3" y="14" width="7" height="7" rx="1" />
              <path stroke-linecap="round" d="M17.5 14v7M14 17.5h7" />
            </svg>
          </div>
          Kiến trúc
          <span
            class="sb-badge"
            style="background: var(--yel-bg); color: var(--yel)"
          >
            {{ getCategoryCount("architecture") }}
          </span>
        </div>
        <div
          class="sb-item"
          @click="toggleFilter('category', 'performance')"
          :class="{ active: activeFilters.category.includes('performance') }"
        >
          <div class="sb-icon" style="color: var(--blu)">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
          </div>
          Hiệu năng
          <span
            class="sb-badge"
            style="background: var(--blu-bg); color: var(--blu)"
          >
            {{ getCategoryCount("performance") }}
          </span>
        </div>
      </div>

      <div class="sb-divider"></div>

      <div class="sb-footer">
        <div class="ai-status-pill">
          <div class="ai-status-dot"></div>
          <div>
            <div class="ai-status-text">Bác sĩ AI sẵn sàng</div>
            <div class="ai-status-model">qwen2.5-coder:3b · Q4_K_M</div>
          </div>
        </div>
      </div>
    </aside>

    <!-- ════════════════════════════════
         MAIN CONTENT
    ════════════════════════════════ -->
    <main class="main">
      <div
        v-if="error && !scanResult"
        class="m-6 p-4 bg-red-bg border border-red text-red rounded-md"
      >
        {{ error }}
        <button class="btn btn-primary mt-2" @click="loadScanData">
          Thử lại
        </button>
      </div>

      <template v-else-if="scanResult">
        <SummaryBar :summary="scanResult.summary" />

        <div class="section-hdr">
          <div class="section-title">
            <div class="section-title-icon">
              <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
                />
              </svg>
            </div>
            Danh sách vấn đề
            <span class="section-count"
              >{{ filteredFindings.length }} findings</span
            >
          </div>
        </div>

        <FilterBar
          :filters="activeFilters"
          :summary="scanResult.summary"
          @update:filters="updateFilters"
        />

        <div class="findings-wrap">
          <div
            v-if="filteredFindings.length === 0"
            class="text-center py-10 text-text3 text-sm"
          >
            Không có findings nào khớp với bộ lọc.
          </div>

          <template v-for="group in groupedFindings" :key="group.severity">
            <div class="group-row" v-if="group.items.length > 0">
              <div
                class="group-badge"
                :class="getSeverityGroupClass(group.severity)"
              >
                <svg
                  v-if="group.severity === 'critical'"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
                  />
                </svg>
                <svg
                  v-else-if="group.severity === 'high'"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2.5"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                <svg
                  v-else-if="group.severity === 'medium'"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <svg
                  v-else
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {{
                  group.severity.charAt(0).toUpperCase() +
                  group.severity.slice(1)
                }}
                &middot; {{ group.items.length }} vấn đề
              </div>
              <div class="group-desc">
                {{ getSeverityDesc(group.severity) }}
              </div>
              <div class="group-line"></div>
            </div>

            <FindingCard
              v-for="(finding, idx) in group.items"
              :key="finding.id"
              :finding="finding"
              :default-open="group.severity === 'critical' && idx === 0"
              @explain="requestAIExplanation"
            />
          </template>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import FilterBar from "./components/FilterBar.vue";
import FindingCard from "./components/FindingCard.vue";
import HealthScore from "./components/HealthScore.vue";
import SummaryBar from "./components/SummaryBar.vue";
import { useScanStore } from "./stores/scan";

const store = useScanStore();

const loading = computed(() => store.loading);
const error = computed(() => store.error);
const scanResult = computed(() => store.scanResult);
const activeReportId = computed(() => store.activeReportId);
const selectedFinding = ref(null);
const sortBy = ref("severity");

const isDark = ref(true);

const activeFilters = ref({
  severity: [],
  category: [],
  search: "",
});

const loadScanData = () => store.loadScanData();

const handleRefresh = async () => {
  await store.refreshToLatest();
};

const updateFilters = (filters) => {
  activeFilters.value = filters;
};

const toggleFilter = (type, value) => {
  const index = activeFilters.value[type].indexOf(value);
  if (index === -1) {
    activeFilters.value[type].push(value);
  } else {
    activeFilters.value[type].splice(index, 1);
  }
};

const getCategoryCount = (cat) => {
  if (!scanResult.value) return 0;
  return scanResult.value.findings.filter((f) => f.category === cat).length;
};

const toggleTheme = () => {
  isDark.value = !isDark.value;
  if (isDark.value) {
    document.documentElement.classList.add("dark");
    localStorage.setItem("theme", "dark");
  } else {
    document.documentElement.classList.remove("dark");
    localStorage.setItem("theme", "light");
  }
};

const getSeverityGroupClass = (sev) => {
  if (sev === "critical") return "gb-red";
  if (sev === "high") return "gb-ora";
  if (sev === "medium") return "gb-yel";
  return "gb-blu";
};

const getSeverityDesc = (sev) => {
  if (sev === "critical") return "Xử lý ngay trước khi deploy production";
  if (sev === "high") return "Xử lý trong tuần này";
  if (sev === "medium") return "Xử lý trong sprint tiếp theo";
  return "Tối ưu khi có thời gian";
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

const groupedFindings = computed(() => {
  const groups = {
    critical: [],
    high: [],
    medium: [],
    low: [],
    info: [],
  };

  filteredFindings.value.forEach((f) => {
    if (groups[f.severity]) {
      groups[f.severity].push(f);
    } else {
      groups.info.push(f);
    }
  });

  return [
    { severity: "critical", items: groups.critical },
    { severity: "high", items: groups.high },
    { severity: "medium", items: groups.medium },
    { severity: "low", items: groups.low },
    { severity: "info", items: groups.info },
  ].filter((g) => g.items.length > 0);
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

onMounted(async () => {
  // Load theme preference
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme === "light") {
    isDark.value = false;
    document.documentElement.classList.remove("dark");
  } else {
    isDark.value = true;
    document.documentElement.classList.add("dark");
  }

  await store.loadScanData();
});
</script>

<style>
/* Layout Shell Styles */
.app {
  display: grid;
  grid-template-columns: 240px 1fr;
  grid-template-rows: 52px 1fr;
  height: 100vh;
  overflow: hidden;
}

/* TOPBAR */
.topbar {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  gap: 0;
  background: var(--bg1);
  border-bottom: 1px solid var(--bd);
  padding: 0;
  z-index: 50;
  position: relative;
}

.topbar::after {
  content: "";
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    var(--acc) 30%,
    var(--acc) 70%,
    transparent
  );
  opacity: 0.2;
}

.tb-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  height: 100%;
  border-right: 1px solid var(--bd);
  min-width: 240px;
  text-decoration: none;
}

.logo-icon {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, var(--acc) 0%, #00a896 100%);
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 0 16px rgba(0, 229, 200, 0.3);
}
.logo-icon svg {
  width: 16px;
  height: 16px;
  color: #0d1117;
}
.logo-name {
  font-family: var(--mono);
  font-size: 13px;
  font-weight: 700;
  color: var(--text0);
  letter-spacing: 0.04em;
}
.logo-tag {
  font-family: var(--mono);
  font-size: 9px;
  color: var(--acc);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  margin-top: -2px;
}

.tb-divider {
  width: 1px;
  height: 24px;
  background: var(--bd2);
  flex-shrink: 0;
}

.tb-project {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}
.tb-project-icon {
  width: 22px;
  height: 22px;
  border-radius: 5px;
  background: var(--bg3);
  border: 1px solid var(--bd2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text2);
  flex-shrink: 0;
}
.tb-proj-name {
  font-size: 13px;
  color: var(--text0);
  font-weight: 500;
}
.tb-proj-sub {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text3);
}

.tb-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: var(--r-sm);
  background: rgba(61, 220, 151, 0.07);
  border: 1px solid rgba(61, 220, 151, 0.18);
  margin: 0 8px;
}
.tb-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--grn);
  box-shadow: 0 0 6px var(--grn);
  animation: pulse-dot 2.5s ease-in-out infinite;
}
.tb-status-text {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--grn);
}

.tb-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  padding-right: 16px;
}

/* SIDEBAR */
.sidebar {
  background: var(--bg1);
  border-right: 1px solid var(--bd);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
}

.sb-section {
  padding: 16px 0 8px;
}
.sb-label {
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text3);
  padding: 0 16px;
  margin-bottom: 4px;
}
.sb-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  cursor: pointer;
  transition: all 0.12s;
  border-left: 2px solid transparent;
  border-radius: 0 6px 6px 0;
  margin: 1px 8px 1px 0;
  font-size: 13px;
  color: var(--text2);
  position: relative;
}
.sb-item:hover {
  background: var(--bg2);
  color: var(--text1);
}
.sb-item.active {
  background: var(--acc-bg);
  border-left-color: var(--acc);
  color: var(--text0);
}
.sb-item.active .sb-icon {
  color: var(--acc) !important;
}

.sb-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--text3);
}
.sb-icon svg {
  width: 15px;
  height: 15px;
}

.sb-badge {
  margin-left: auto;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  padding: 1px 7px;
  border-radius: 99px;
  flex-shrink: 0;
}
.sb-divider {
  height: 1px;
  background: var(--bd);
  margin: 6px 16px;
}

.sb-footer {
  margin-top: auto;
  padding: 12px;
  border-top: 1px solid var(--bd);
}
.ai-status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: var(--bg2);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
  font-size: 11px;
  color: var(--text2);
}
.ai-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--grn);
  box-shadow: 0 0 5px var(--grn);
  flex-shrink: 0;
}
.ai-status-text {
  flex: 1;
  font-weight: 500;
}
.ai-status-model {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text3);
}

/* MAIN CONTENT */
.main {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  background: var(--bg0);
}

.section-hdr {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 10px;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text0);
}
.section-title-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: var(--acc-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--acc);
}
.section-title-icon svg {
  width: 13px;
  height: 13px;
}
.section-count {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text2);
  background: var(--bg3);
  padding: 1px 8px;
  border-radius: 99px;
  border: 1px solid var(--bd);
}

.findings-wrap {
  padding: 0 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.group-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0 6px;
  margin-top: 6px;
}
.group-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 99px;
  letter-spacing: 0.04em;
  white-space: nowrap;
}
.group-badge svg {
  width: 11px;
  height: 11px;
}

.gb-red {
  background: var(--red-bg);
  color: var(--red);
  border: 1px solid rgba(239, 68, 68, 0.2);
}
.gb-ora {
  background: var(--ora-bg);
  color: var(--ora);
  border: 1px solid rgba(249, 115, 22, 0.2);
}
.gb-yel {
  background: var(--yel-bg);
  color: var(--yel);
  border: 1px solid rgba(234, 179, 8, 0.2);
}
.gb-blu {
  background: var(--blu-bg);
  color: var(--blu);
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.group-desc {
  font-size: 11px;
  color: var(--text3);
  white-space: nowrap;
}
.group-line {
  flex: 1;
  height: 1px;
  background: var(--bd);
}
</style>
