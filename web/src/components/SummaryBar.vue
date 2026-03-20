<template>
  <div class="overview">
    <div class="stat-card c-red">
      <div class="stat-top">
        <div class="stat-icon">
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
      </div>
      <div class="stat-num">{{ summary.critical || 0 }}</div>
      <div class="stat-label">Critical</div>
    </div>

    <div class="stat-card c-ora">
      <div class="stat-top">
        <div class="stat-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
        </div>
      </div>
      <div class="stat-num">{{ summary.high || 0 }}</div>
      <div class="stat-label">High</div>
    </div>

    <div class="stat-card c-yel">
      <div class="stat-top">
        <div class="stat-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
      </div>
      <div class="stat-num">{{ summary.medium || 0 }}</div>
      <div class="stat-label">Medium</div>
    </div>

    <div class="stat-card c-blu">
      <div class="stat-top">
        <div class="stat-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
        </div>
      </div>
      <div class="stat-num">{{ summary.low || 0 }}</div>
      <div class="stat-label">Low</div>
    </div>

    <div class="stat-card c-text">
      <div class="stat-top">
        <div class="stat-icon">
          <svg
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
      </div>
      <div class="stat-num" style="color: var(--text0)">
        {{ summary.total || 0 }}
      </div>
      <div class="stat-label">Tổng vấn đề</div>
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
      total: 0,
    }),
  },
});
</script>

<style scoped>
.overview {
  padding: 20px 24px 0;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
}

.stat-card {
  background: var(--bg1);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
  padding: 14px 16px;
  position: relative;
  overflow: hidden;
  transition: border-color 0.2s;
  cursor: default;
}

.stat-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  border-radius: var(--r-md) var(--r-md) 0 0;
}

.stat-card.c-red::before {
  background: linear-gradient(90deg, var(--red), transparent);
}
.stat-card.c-ora::before {
  background: linear-gradient(90deg, var(--ora), transparent);
}
.stat-card.c-yel::before {
  background: linear-gradient(90deg, var(--yel), transparent);
}
.stat-card.c-blu::before {
  background: linear-gradient(90deg, var(--blu), transparent);
}
.stat-card.c-text::before {
  background: linear-gradient(90deg, var(--text3), transparent);
}

.stat-card:hover {
  border-color: var(--bd3);
}

.stat-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.stat-icon {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon svg {
  width: 14px;
  height: 14px;
}

.stat-card.c-red .stat-icon {
  background: var(--red-bg);
  color: var(--red);
}
.stat-card.c-ora .stat-icon {
  background: var(--ora-bg);
  color: var(--ora);
}
.stat-card.c-yel .stat-icon {
  background: var(--yel-bg);
  color: var(--yel);
}
.stat-card.c-blu .stat-icon {
  background: var(--blu-bg);
  color: var(--blu);
}
.stat-card.c-text .stat-icon {
  background: var(--bg3);
  color: var(--text2);
}

.stat-num {
  font-family: var(--mono);
  font-size: 26px;
  font-weight: 700;
  line-height: 1;
  color: var(--text0);
  margin-bottom: 2px;
}
.stat-card.c-red .stat-num {
  color: var(--red);
}

.stat-label {
  font-size: 11px;
  color: var(--text2);
  font-weight: 500;
}
</style>
