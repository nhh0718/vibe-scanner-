<template>
  <div
    class="bg-vs-card border border-vs-border rounded-xl p-5 hover:border-vs-primary/50 transition-colors"
  >
    <!-- Header -->
    <div class="flex items-start justify-between mb-3">
      <div class="flex items-center space-x-3">
        <span class="font-mono text-sm text-vs-primary"
          >[{{ finding.id }}]</span
        >
        <span
          class="px-2 py-1 rounded-full text-xs font-medium"
          :class="getSeverityClass(finding.severity)"
        >
          {{ finding.severity }}
        </span>
        <span
          class="px-2 py-1 rounded text-xs"
          :class="getCategoryClass(finding.category)"
        >
          {{ getCategoryLabel(finding.category) }}
        </span>
        <span
          v-if="finding.engine"
          class="px-2 py-1 rounded text-xs bg-gray-700/50 text-gray-300 border border-gray-600/30"
          :title="'Detected by: ' + finding.engine"
        >
          {{ getEngineLabel(finding.engine) }}
        </span>
      </div>
      <button
        @click="toggleExpand"
        class="text-vs-text-muted hover:text-vs-primary transition-colors"
      >
        {{ expanded ? "▲" : "▼" }}
      </button>
    </div>

    <!-- Title & File -->
    <h3 class="text-lg font-medium text-vs-text mb-2">{{ finding.title }}</h3>
    <div class="flex items-center space-x-2 text-sm mb-3">
      <span class="text-vs-primary font-mono"
        >{{ finding.file }}:{{ finding.line }}</span
      >
    </div>

    <!-- Message -->
    <p class="text-vs-text-muted mb-3">{{ finding.message }}</p>

    <!-- Code Snippet -->
    <div v-if="finding.code_snippet" v-show="expanded" class="mb-4">
      <div class="bg-vs-darker rounded-lg p-4 overflow-x-auto">
        <pre class="text-sm text-vs-text font-mono">{{
          finding.code_snippet
        }}</pre>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center space-x-3">
      <button
        @click="requestExplanation"
        class="flex items-center space-x-2 px-4 py-2 bg-vs-secondary/10 text-vs-secondary rounded-lg hover:bg-vs-secondary/20 transition-colors"
        :disabled="loading"
      >
        <span v-if="loading" class="animate-spin">⏳</span>
        <span v-else>🤖</span>
        <span>{{ loading ? "Đang hỏi AI..." : "Hỏi bác sĩ AI" }}</span>
      </button>

      <a
        v-if="finding.references && finding.references.length > 0"
        :href="finding.references[0]"
        target="_blank"
        class="text-sm text-vs-text-muted hover:text-vs-primary transition-colors"
      >
        📚 Tài liệu tham khảo
      </a>
    </div>

    <!-- AI Explanation -->
    <div
      v-if="explanation"
      class="mt-4 p-4 bg-vs-darker/50 rounded-lg border border-vs-border"
    >
      <div class="flex items-center space-x-2 mb-2">
        <span class="text-vs-secondary">🤖</span>
        <span class="font-medium text-vs-secondary">Giải thích từ AI:</span>
      </div>
      <div class="text-vs-text whitespace-pre-line">{{ explanation }}</div>
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
});

const emit = defineEmits(["explain"]);

const expanded = ref(true);
const loading = ref(false);
const explanation = ref(props.finding.explanation || "");

const toggleExpand = () => {
  expanded.value = !expanded.value;
};

const requestExplanation = async () => {
  if (explanation.value) {
    // Toggle visibility if already has explanation
    return;
  }

  loading.value = true;

  try {
    const response = await fetch(`/api/ai/explain/${props.finding.id}`);
    if (response.ok) {
      const data = await response.json();
      explanation.value = data.explanation;
    } else {
      // Fallback: emit to parent for modal display
      emit("explain", props.finding);
    }
  } catch (err) {
    emit("explain", props.finding);
  } finally {
    loading.value = false;
  }
};

const getSeverityClass = (severity) => {
  const classes = {
    critical: "bg-vs-danger/20 text-vs-danger border border-vs-danger/30",
    high: "bg-orange-500/20 text-orange-500 border border-orange-500/30",
    medium: "bg-vs-warning/20 text-vs-warning border border-vs-warning/30",
    low: "bg-vs-info/20 text-vs-info border border-vs-info/30",
    info: "bg-gray-500/20 text-gray-500 border border-gray-500/30",
  };
  return classes[severity] || classes.info;
};

const getCategoryClass = (category) => {
  const classes = {
    security: "bg-vs-danger/10 text-vs-danger",
    quality: "bg-vs-info/10 text-vs-info",
    architecture: "bg-vs-secondary/10 text-vs-secondary",
    secrets: "bg-pink-500/10 text-pink-500",
    performance: "bg-vs-success/10 text-vs-success",
  };
  return classes[category] || "bg-gray-500/10 text-gray-500";
};

const getCategoryLabel = (category) => {
  const labels = {
    security: "🔐 Bảo mật",
    quality: "✨ Chất lượng",
    architecture: "🏗️ Kiến trúc",
    secrets: "🔑 Secrets",
    performance: "⚡ Hiệu năng",
  };
  return labels[category] || category;
};

const getEngineLabel = (engine) => {
  const labels = {
    semgrep: "🔍 Semgrep",
    gitleaks: "🔐 Gitleaks",
    "native-security": "🛡️ Native",
    vibesecurity: "🌳 AST",
    ast: "🌳 AST",
    complexity: "📊 Complex",
    "dependency-audit": "📦 Deps",
  };
  return labels[engine] || engine;
};
</script>
