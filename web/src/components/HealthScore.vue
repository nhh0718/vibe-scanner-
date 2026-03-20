<template>
  <div class="sb-health">
    <div class="sb-health-top">
      <div>
        <div
          style="
            font-family: var(--mono);
            font-size: 10px;
            color: var(--text3);
            margin-bottom: 4px;
            letter-spacing: 0.06em;
          "
        >
          ĐIỂM SỨC KHỎE
        </div>
        <div
          class="sb-health-score"
          :style="{ color: getOverallColor(score.overall) }"
        >
          {{ score.overall || 0 }}
        </div>
        <div
          style="font-family: var(--mono); font-size: 9px; color: var(--text3)"
        >
          /100
        </div>
      </div>
      <div class="sb-health-label" :class="getOverallLabelClass(score.overall)">
        {{ getStatusText(score.overall) }}
      </div>
    </div>
    <div class="sb-mini-bars">
      <div class="sb-mini-bar-row">
        <div class="sb-mini-bar-label">Bảo mật</div>
        <div class="sb-mini-bar-track">
          <div
            class="sb-mini-bar-fill"
            :style="{
              width: `${score.security || 0}%`,
              background: 'var(--red)',
            }"
          ></div>
        </div>
        <div class="sb-mini-bar-val" style="color: var(--red)">
          {{ score.security || 0 }}
        </div>
      </div>
      <div class="sb-mini-bar-row">
        <div class="sb-mini-bar-label">Chất lượng</div>
        <div class="sb-mini-bar-track">
          <div
            class="sb-mini-bar-fill"
            :style="{
              width: `${score.quality || 0}%`,
              background: 'var(--ora)',
            }"
          ></div>
        </div>
        <div class="sb-mini-bar-val" style="color: var(--ora)">
          {{ score.quality || 0 }}
        </div>
      </div>
      <div class="sb-mini-bar-row">
        <div class="sb-mini-bar-label">Kiến trúc</div>
        <div class="sb-mini-bar-track">
          <div
            class="sb-mini-bar-fill"
            :style="{
              width: `${score.architecture || 0}%`,
              background: 'var(--yel)',
            }"
          ></div>
        </div>
        <div class="sb-mini-bar-val" style="color: var(--yel)">
          {{ score.architecture || 0 }}
        </div>
      </div>
      <div class="sb-mini-bar-row">
        <div class="sb-mini-bar-label">Hiệu năng</div>
        <div class="sb-mini-bar-track">
          <div
            class="sb-mini-bar-fill"
            :style="{
              width: `${score.performance || 0}%`,
              background: 'var(--blu)',
            }"
          ></div>
        </div>
        <div class="sb-mini-bar-val" style="color: var(--blu)">
          {{ score.performance || 0 }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  score: {
    type: Object,
    required: true,
    default: () => ({
      overall: 0,
      security: 0,
      quality: 0,
      architecture: 0,
      performance: 0,
    }),
  },
});

const getOverallColor = (score) => {
  if (score >= 80) return "var(--grn)";
  if (score >= 60) return "var(--yel)";
  if (score >= 40) return "var(--ora)";
  return "var(--red)";
};

const getOverallLabelClass = (score) => {
  if (score >= 80) return "hl-grn";
  if (score >= 60) return "hl-yel";
  if (score >= 40) return "hl-ora";
  return "hl-red";
};

const getStatusText = (score) => {
  if (score >= 80) return "TỐT";
  if (score >= 60) return "TRUNG BÌNH";
  if (score >= 40) return "CẢNH BÁO";
  return "KHẨN CẤP";
};
</script>

<style scoped>
.sb-health {
  padding: 12px;
  background: var(--bg2);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
}

.sb-health-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.sb-health-score {
  font-family: var(--mono);
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.sb-health-label {
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 99px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.sb-health-label.hl-grn {
  background: var(--grn-bg);
  color: var(--grn);
}
.sb-health-label.hl-yel {
  background: var(--yel-bg);
  color: var(--yel);
}
.sb-health-label.hl-ora {
  background: var(--ora-bg);
  color: var(--ora);
}
.sb-health-label.hl-red {
  background: var(--red-bg);
  color: var(--red);
}

.sb-mini-bars {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.sb-mini-bar-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.sb-mini-bar-label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text3);
  width: 60px;
  flex-shrink: 0;
}
.sb-mini-bar-track {
  flex: 1;
  height: 3px;
  background: var(--bg4);
  border-radius: 99px;
  overflow: hidden;
}
.sb-mini-bar-fill {
  height: 100%;
  border-radius: 99px;
  transition: width 1s cubic-bezier(0.4, 0, 0.2, 1);
}
.sb-mini-bar-val {
  font-family: var(--mono);
  font-size: 10px;
  width: 22px;
  text-align: right;
}
</style>
