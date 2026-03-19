<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="close">
    <div class="bg-vs-card border border-vs-border rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-vs-border">
        <div class="flex items-center space-x-3">
          <span class="text-2xl">🤖</span>
          <h2 class="text-xl font-semibold text-vs-text">Giải thích từ bác sĩ AI</h2>
        </div>
        <button @click="close" class="text-vs-text-muted hover:text-vs-text text-2xl">&times;</button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-6">
        <!-- Finding Info -->
        <div class="bg-vs-darker rounded-lg p-4">
          <div class="flex items-center space-x-2 mb-2">
            <span
              class="px-2 py-1 rounded text-xs font-medium"
              :class="getSeverityClass(finding.severity)"
            >
              {{ finding.severity }}
            </span>
            <span class="text-vs-primary font-mono text-sm">{{ finding.file }}:{{ finding.line }}</span>
          </div>
          <h3 class="text-lg font-medium text-vs-text">{{ finding.title }}</h3>
        </div>

        <!-- AI Response -->
        <div v-if="loading" class="text-center py-8">
          <div class="animate-pulse space-y-4">
            <div class="h-4 bg-vs-border rounded w-3/4 mx-auto"></div>
            <div class="h-4 bg-vs-border rounded w-1/2 mx-auto"></div>
            <div class="h-4 bg-vs-border rounded w-2/3 mx-auto"></div>
          </div>
          <p class="text-vs-text-muted mt-4">🤔 AI đang phân tích...</p>
        </div>

        <div v-else-if="error" class="bg-vs-danger/10 border border-vs-danger/30 rounded-lg p-4">
          <p class="text-vs-danger">{{ error }}</p>
          <p class="text-vs-text-muted text-sm mt-2">
            Đảm bảo bạn đã chạy <code class="bg-vs-darker px-2 py-1 rounded">vibescanner ai-setup</code> để cài đặt AI.
          </p>
        </div>

        <div v-else-if="explanation" class="space-y-4">
          <div class="prose prose-invert max-w-none">
            <div class="whitespace-pre-line text-vs-text">{{ explanation }}</div>
          </div>

          <!-- Action Buttons -->
          <div class="flex space-x-3 pt-4 border-t border-vs-border">
            <button
              v-if="fixCode"
              @click="copyFix"
              class="flex-1 px-4 py-2 bg-vs-success/10 text-vs-success rounded-lg hover:bg-vs-success/20 transition-colors"
            >
              📋 Copy code fix
            </button>
            <button
              @click="close"
              class="px-4 py-2 bg-vs-border text-vs-text rounded-lg hover:bg-vs-border/80 transition-colors"
            >
              Đóng
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  finding: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close'])

const loading = ref(true)
const error = ref(null)
const explanation = ref('')
const fixCode = ref('')

onMounted(() => {
  fetchExplanation()
})

const fetchExplanation = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await fetch(`/api/ai/explain/${props.finding.id}`)

    if (!response.ok) {
      if (response.status === 503) {
        throw new Error('AI chưa được cài đặt. Vui lòng chạy "vibescanner ai-setup" trước.')
      }
      throw new Error('Không thể lấy giải thích từ AI')
    }

    const data = await response.json()
    explanation.value = data.explanation
    fixCode.value = data.fix_code || ''
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const copyFix = () => {
  if (fixCode.value) {
    navigator.clipboard.writeText(fixCode.value)
    alert('Đã copy code fix vào clipboard!')
  }
}

const close = () => {
  emit('close')
}

const getSeverityClass = (severity) => {
  const classes = {
    critical: 'bg-vs-danger/20 text-vs-danger',
    high: 'bg-orange-500/20 text-orange-500',
    medium: 'bg-vs-warning/20 text-vs-warning',
    low: 'bg-vs-info/20 text-vs-info',
    info: 'bg-gray-500/20 text-gray-500'
  }
  return classes[severity] || classes.info
}
</script>
