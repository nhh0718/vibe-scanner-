<template>
  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    <!-- Overall Score -->
    <div class="bg-vs-card border border-vs-border rounded-xl p-6 text-center">
      <div
        class="text-4xl font-bold mb-2"
        :class="getScoreColor(score.overall)"
      >
        {{ score.overall }}
      </div>
      <div class="text-sm text-vs-text-muted uppercase tracking-wide">Tổng quát</div>
      <div class="mt-2 text-xs" :class="getStatusColor(score.overall)">
        {{ getStatusText(score.overall) }}
      </div>
    </div>

    <!-- Security Score -->
    <div class="bg-vs-card border border-vs-border rounded-xl p-6 text-center">
      <div
        class="text-4xl font-bold mb-2"
        :class="getScoreColor(score.security)"
      >
        {{ score.security }}
      </div>
      <div class="text-sm text-vs-text-muted uppercase tracking-wide">Bảo mật</div>
      <div class="mt-2 text-xs" :class="getStatusColor(score.security)">
        {{ getStatusText(score.security) }}
      </div>
    </div>

    <!-- Quality Score -->
    <div class="bg-vs-card border border-vs-border rounded-xl p-6 text-center">
      <div
        class="text-4xl font-bold mb-2"
        :class="getScoreColor(score.quality)"
      >
        {{ score.quality }}
      </div>
      <div class="text-sm text-vs-text-muted uppercase tracking-wide">Chất lượng</div>
      <div class="mt-2 text-xs" :class="getStatusColor(score.quality)">
        {{ getStatusText(score.quality) }}
      </div>
    </div>

    <!-- Architecture Score -->
    <div class="bg-vs-card border border-vs-border rounded-xl p-6 text-center">
      <div
        class="text-4xl font-bold mb-2"
        :class="getScoreColor(score.architecture)"
      >
        {{ score.architecture }}
      </div>
      <div class="text-sm text-vs-text-muted uppercase tracking-wide">Kiến trúc</div>
      <div class="mt-2 text-xs" :class="getStatusColor(score.architecture)">
        {{ getStatusText(score.architecture) }}
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
      performance: 0
    })
  }
})

const getScoreColor = (score) => {
  if (score >= 80) return 'text-vs-success'
  if (score >= 60) return 'text-vs-warning'
  if (score >= 40) return 'text-orange-500'
  return 'text-vs-danger'
}

const getStatusColor = (score) => {
  if (score >= 80) return 'text-vs-success'
  if (score >= 60) return 'text-vs-warning'
  if (score >= 40) return 'text-orange-500'
  return 'text-vs-danger'
}

const getStatusText = (score) => {
  if (score >= 80) return 'Tốt'
  if (score >= 60) return 'Trung bình'
  if (score >= 40) return 'Cần cải thiện'
  return 'Nguy hiểm'
}
</script>
