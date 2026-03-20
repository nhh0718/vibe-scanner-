<template>
  <div
    class="fcard"
    :class="[getSeverityClass(finding.severity), { open: expanded }]"
    @click="toggleExpand"
  >
    <div class="fcard-left-bar"></div>
    <div class="fcard-head">
      <span class="fcard-sev">{{ finding.severity.toUpperCase() }}</span>
      <div class="fcard-info">
        <div class="fcard-title">{{ finding.title }}</div>
        <div class="fcard-meta">
          <span class="file">
            <svg
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
              />
            </svg>
            {{ finding.file }} &middot; dòng {{ finding.line }}
          </span>
          <span v-if="finding.id">&middot;</span>
          <span v-if="finding.id">{{ finding.id }}</span>
        </div>
      </div>
      <div class="fcard-tags">
        <span class="tag" v-if="finding.category">{{ finding.category }}</span>
        <span class="tag" v-if="finding.engine">{{ finding.engine }}</span>
      </div>
      <div class="fcard-chevron">
        <svg
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </div>
    </div>

    <div class="fcard-body" @click.stop>
      <div class="fcard-body-inner">
        <!-- Cột trái: Vấn đề & Cách sửa (nếu có message text) -->
        <div class="fbcol">
          <div class="fb-label">
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
            Vấn đề
          </div>
          <div class="fb-desc">
            {{ finding.message }}
          </div>

          <div v-if="finding.references && finding.references.length > 0">
            <div class="fb-label" style="margin-top: 14px">
              <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                />
              </svg>
              Tài liệu tham khảo
            </div>
            <a
              :href="finding.references[0]"
              target="_blank"
              class="text-[12px] text-acc hover:underline mt-1 inline-block"
            >
              {{ finding.references[0] }}
            </a>
          </div>
        </div>

        <!-- Cột phải: Code & AI -->
        <div class="fbcol">
          <div class="fb-label">
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
            Đoạn code có lỗi
          </div>

          <div class="code-snippet" v-if="finding.code_snippet">
            <div class="snippet-hdr">
              <span class="snippet-filename">
                <svg
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                  />
                </svg>
                {{ finding.file }}
              </span>
              <span style="color: var(--red); font-size: 10px"
                >&bull; dòng {{ finding.line }}</span
              >
            </div>
            <div class="snippet-body overflow-x-auto">
              <pre class="px-3 text-[11px] text-text1 m-0 leading-relaxed">{{
                finding.code_snippet
              }}</pre>
            </div>
          </div>
          <div v-else class="text-xs text-text3 italic">
            Không có code snippet
          </div>

          <div class="ai-section">
            <button
              class="ai-btn"
              :class="{ active: showAI }"
              @click="requestExplanation"
              :disabled="loading"
            >
              <svg
                :class="{ 'animate-spin': loading }"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  v-if="loading"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M4 4v5h5M20 20v-5h-5M4 9a9 9 0 0115.457-5.457M20 15a9 9 0 01-15.457 5.457l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
                <path
                  v-else
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
              {{
                loading ? "Đang hỏi AI..." : "Hỏi bác sĩ AI — giải thích thêm"
              }}
            </button>

            <div class="ai-panel" :class="{ vis: showAI }">
              <div class="ai-panel-hdr">
                <svg
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                  />
                </svg>
                Bác sĩ AI
              </div>
              <div class="ai-panel-body whitespace-pre-line">
                {{ explanation }}
                <span v-if="loading" class="ai-cursor"></span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";

const props = defineProps({
  finding: {
    type: Object,
    required: true,
  },
  defaultOpen: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["explain"]);

const expanded = ref(props.defaultOpen);
const showAI = ref(!!props.finding.explanation);
const loading = ref(false);
const explanation = ref(props.finding.explanation || "");

const toggleExpand = () => {
  expanded.value = !expanded.value;
};

const requestExplanation = async () => {
  if (explanation.value) {
    showAI.value = !showAI.value;
    return;
  }

  loading.value = true;
  showAI.value = true;
  explanation.value = "";

  try {
    const response = await fetch(`/api/ai/explain/${props.finding.id}`);
    if (response.ok) {
      const data = await response.json();
      explanation.value = data.explanation;
    } else {
      explanation.value =
        "Lỗi khi gọi API AI. Hãy chắc chắn bạn đã cấu hình AI bằng 'vibescanner ai-setup'.";
    }
  } catch (err) {
    explanation.value = "Lỗi kết nối tới AI: " + err.message;
  } finally {
    loading.value = false;
  }
};

const getSeverityClass = (severity) => {
  const classes = {
    critical: "sev-c",
    high: "sev-h",
    medium: "sev-m",
    low: "sev-l",
    info: "sev-i",
  };
  return classes[severity] || classes.info;
};
</script>

<style scoped>
.fcard {
  background: var(--bg1);
  border: 1px solid var(--bd);
  border-radius: var(--r-md);
  margin-bottom: 4px;
  overflow: hidden;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
  cursor: pointer;
  position: relative;
  animation: slideUp 0.3s both;
}

.fcard:hover {
  border-color: var(--bd3);
}
.fcard.open {
  border-color: var(--bd2);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.fcard-left-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: var(--r-md) 0 0 var(--r-md);
  background: var(--text3);
}
.fcard.sev-c .fcard-left-bar {
  background: var(--red);
}
.fcard.sev-h .fcard-left-bar {
  background: var(--ora);
}
.fcard.sev-m .fcard-left-bar {
  background: var(--yel);
}
.fcard.sev-l .fcard-left-bar {
  background: var(--blu);
}

.fcard-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px 12px 18px;
  user-select: none;
}

.fcard-sev {
  font-family: var(--mono);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.1em;
  padding: 3px 8px;
  border-radius: 99px;
  white-space: nowrap;
  flex-shrink: 0;
  background: var(--bg3);
  color: var(--text2);
  border: 1px solid var(--bd);
}
.fcard.sev-c .fcard-sev {
  background: var(--red-bg);
  color: var(--red);
  border-color: rgba(239, 68, 68, 0.25);
}
.fcard.sev-h .fcard-sev {
  background: var(--ora-bg);
  color: var(--ora);
  border-color: rgba(249, 115, 22, 0.25);
}
.fcard.sev-m .fcard-sev {
  background: var(--yel-bg);
  color: var(--yel);
  border-color: rgba(234, 179, 8, 0.25);
}
.fcard.sev-l .fcard-sev {
  background: var(--blu-bg);
  color: var(--blu);
  border-color: rgba(59, 130, 246, 0.25);
}

.fcard-info {
  flex: 1;
  min-width: 0;
}
.fcard-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text0);
  margin-bottom: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fcard-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text3);
}
.fcard-meta .file {
  color: var(--acc2);
  display: flex;
  align-items: center;
  gap: 4px;
}
.fcard-meta .file svg {
  width: 10px;
  height: 10px;
}

.fcard-tags {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
.tag {
  font-family: var(--mono);
  font-size: 10px;
  padding: 1px 7px;
  border-radius: 4px;
  background: var(--bg3);
  color: var(--text3);
  border: 1px solid var(--bd);
}

.fcard-chevron {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text3);
  border-radius: 6px;
  transition: all 0.2s;
  flex-shrink: 0;
  background: var(--bg3);
  border: 1px solid var(--bd);
}
.fcard-chevron svg {
  width: 12px;
  height: 12px;
  transition: transform 0.25s;
}
.fcard.open .fcard-chevron svg {
  transform: rotate(180deg);
}
.fcard.open .fcard-chevron {
  color: var(--text0);
  border-color: var(--bd2);
}

.fcard-body {
  display: none;
  border-top: 1px solid var(--bd);
  cursor: default;
}
.fcard.open .fcard-body {
  display: block;
}

.fcard-body-inner {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
}
.fbcol {
  padding: 16px 18px;
}
.fbcol + .fbcol {
  border-left: 1px solid var(--bd);
}

@media (max-width: 1024px) {
  .fcard-body-inner {
    grid-template-columns: 1fr;
  }
  .fbcol + .fbcol {
    border-left: none;
    border-top: 1px solid var(--bd);
  }
}

.fb-label {
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: var(--text3);
  text-transform: uppercase;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.fb-label svg {
  width: 11px;
  height: 11px;
}

.fb-desc {
  font-size: 13px;
  color: var(--text1);
  line-height: 1.7;
  margin-bottom: 14px;
}

/* Snippet */
.code-snippet {
  background: var(--bg0);
  border: 1px solid var(--bd);
  border-radius: var(--r-sm);
  overflow: hidden;
  font-family: var(--mono);
  font-size: 11.5px;
  line-height: 1.8;
}
.snippet-hdr {
  background: var(--bg2);
  padding: 5px 12px;
  font-size: 10px;
  color: var(--text3);
  border-bottom: 1px solid var(--bd);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.snippet-filename {
  color: var(--acc2);
  display: flex;
  align-items: center;
  gap: 5px;
}
.snippet-filename svg {
  width: 11px;
  height: 11px;
}
.snippet-body {
  padding: 8px 0;
}

/* AI */
.ai-section {
  margin-top: 14px;
}
.ai-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 9px 16px;
  border-radius: var(--r-md);
  background: none;
  border: 1px dashed rgba(13, 148, 136, 0.25);
  color: var(--acc);
  font-family: var(--sans);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.dark .ai-btn {
  border-color: rgba(0, 229, 200, 0.25);
}

.ai-btn:hover,
.ai-btn.active {
  background: var(--acc-bg);
  border-style: solid;
  border-color: rgba(13, 148, 136, 0.4);
}
.dark .ai-btn:hover,
.dark .ai-btn.active {
  border-color: rgba(0, 229, 200, 0.4);
}

.ai-btn svg {
  width: 14px;
  height: 14px;
}

.ai-panel {
  display: none;
  margin-top: 10px;
  background: var(--bg0);
  border: 1px solid rgba(13, 148, 136, 0.18);
  border-radius: var(--r-md);
  overflow: hidden;
}
.dark .ai-panel {
  border-color: rgba(0, 229, 200, 0.18);
}

.ai-panel.vis {
  display: block;
}
.ai-panel-hdr {
  padding: 8px 12px;
  background: var(--acc-bg);
  border-bottom: 1px solid var(--acc-bg2);
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: var(--mono);
  font-size: 10px;
  color: var(--acc);
  letter-spacing: 0.08em;
}
.ai-panel-hdr svg {
  width: 12px;
  height: 12px;
}

.ai-cursor {
  display: inline-block;
  width: 7px;
  height: 13px;
  background: var(--acc);
  vertical-align: middle;
  margin-left: 2px;
  border-radius: 1px;
  animation: blink 0.85s step-end infinite;
}

.ai-panel-body {
  padding: 12px;
  font-size: 12.5px;
  line-height: 1.75;
  color: var(--text1);
}
</style>
