<template>
  <div class="min-h-screen bg-vs-dark">
    <!-- Header -->
    <header class="bg-vs-card border-b border-vs-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="text-3xl">🔍</div>
            <div>
              <h1 class="text-2xl font-bold bg-gradient-to-r from-vs-primary to-vs-secondary bg-clip-text text-transparent">
                VibeScanner Dashboard
              </h1>
              <p v-if="scanResult" class="text-sm text-vs-text-muted">
                {{ scanResult.project.name }} • {{ formatDate(scanResult.timestamp) }}
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-4">
            <button
              @click="refreshScan"
              class="px-4 py-2 bg-vs-primary/10 text-vs-primary rounded-lg hover:bg-vs-primary/20 transition-colors"
            >
              🔄 Làm mới
            </button>
            <div class="text-xs text-vs-text-muted bg-vs-darker px-3 py-1 rounded-full">
              🔒 100% Local
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vs-primary mx-auto mb-4"></div>
          <p class="text-vs-text-muted">Đang tải dữ liệu scan...</p>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="bg-vs-danger/10 border border-vs-danger/30 rounded-xl p-6 text-center">
        <p class="text-vs-danger">{{ error }}</p>
        <button @click="loadScanData" class="mt-4 px-4 py-2 bg-vs-danger/20 text-vs-danger rounded-lg">
          Thử lại
        </button>
      </div>

      <!-- Dashboard Content -->
      <div v-else-if="scanResult" class="space-y-6">
        <!-- Health Score Cards -->
        <HealthScore :score="scanResult.health_score" />

        <!-- Summary Bar -->
        <SummaryBar :summary="scanResult.summary" />

        <!-- Filters -->
        <FilterBar
          :filters="activeFilters"
          @update:filters="updateFilters"
        />

        <!-- Findings List -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-semibold text-vs-text">
              🚨 Phát hiện ({{ filteredFindings.length }})
            </h2>
            <div class="flex items-center space-x-2">
              <span class="text-sm text-vs-text-muted">Sắp xếp:</span>
              <select
                v-model="sortBy"
                class="bg-vs-card border border-vs-border rounded-lg px-3 py-1 text-sm text-vs-text"
              >
                <option value="severity">Mức độ nghiêm trọng</option>
                <option value="category">Danh mục</option>
                <option value="file">File</option>
              </select>
            </div>
          </div>

          <div v-if="filteredFindings.length === 0" class="text-center py-12 bg-vs-card rounded-xl">
            <p class="text-vs-text-muted">Không có findings nào khớp với bộ lọc 🎉</p>
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
import { ref, computed, onMounted } from 'vue'
import HealthScore from './components/HealthScore.vue'
import SummaryBar from './components/SummaryBar.vue'
import FilterBar from './components/FilterBar.vue'
import FindingCard from './components/FindingCard.vue'
import AIModal from './components/AIModal.vue'

const loading = ref(true)
const error = ref(null)
const scanResult = ref(null)
const selectedFinding = ref(null)
const sortBy = ref('severity')

const activeFilters = ref({
  severity: [],
  category: [],
  search: ''
})

const loadScanData = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await fetch('/api/scan')
    if (!response.ok) {
      throw new Error('Không thể tải dữ liệu scan')
    }
    scanResult.value = await response.json()
  } catch (err) {
    error.value = err.message
    // Fallback: generate sample data for development
    scanResult.value = generateSampleData()
  } finally {
    loading.value = false
  }
}

const refreshScan = () => {
  loadScanData()
}

const updateFilters = (filters) => {
  activeFilters.value = filters
}

const filteredFindings = computed(() => {
  if (!scanResult.value) return []

  let findings = [...scanResult.value.findings]

  // Apply filters
  if (activeFilters.value.severity.length > 0) {
    findings = findings.filter(f => activeFilters.value.severity.includes(f.severity))
  }

  if (activeFilters.value.category.length > 0) {
    findings = findings.filter(f => activeFilters.value.category.includes(f.category))
  }

  if (activeFilters.value.search) {
    const search = activeFilters.value.search.toLowerCase()
    findings = findings.filter(f =>
      f.title.toLowerCase().includes(search) ||
      f.message.toLowerCase().includes(search) ||
      f.file.toLowerCase().includes(search)
    )
  }

  // Sort
  const severityOrder = { critical: 0, high: 1, medium: 2, low: 3, info: 4 }

  switch (sortBy.value) {
    case 'severity':
      findings.sort((a, b) => severityOrder[a.severity] - severityOrder[b.severity])
      break
    case 'category':
      findings.sort((a, b) => a.category.localeCompare(b.category))
      break
    case 'file':
      findings.sort((a, b) => a.file.localeCompare(b.file))
      break
  }

  return findings
})

const requestAIExplanation = (finding) => {
  selectedFinding.value = finding
}

const formatDate = (timestamp) => {
  return new Date(timestamp).toLocaleString('vi-VN')
}

// Sample data for development
const generateSampleData = () => {
  return {
    scan_id: 'sample-123',
    timestamp: new Date().toISOString(),
    project: {
      name: 'Sample Project',
      path: '/path/to/project',
      languages: ['javascript', 'typescript'],
      files_scanned: 42
    },
    health_score: {
      overall: 65,
      security: 45,
      quality: 70,
      architecture: 60,
      performance: 80
    },
    summary: {
      critical: 2,
      high: 5,
      medium: 12,
      low: 20,
      info: 8,
      total: 47
    },
    findings: [
      {
        id: 'F-001',
        rule_id: 'sql-injection',
        severity: 'critical',
        category: 'security',
        title: 'SQL Injection vulnerability',
        message: 'User input is directly concatenated into SQL query',
        file: 'src/db/user.js',
        line: 45,
        code_snippet: 'const query = "SELECT * FROM users WHERE id = " + userId'
      },
      {
        id: 'F-002',
        rule_id: 'hardcoded-secret',
        severity: 'critical',
        category: 'secrets',
        title: 'Hardcoded API key detected',
        message: 'API key found in source code',
        file: 'src/config.js',
        line: 12,
        code_snippet: 'const API_KEY = "sk-1234567890abcdef"'
      },
      {
        id: 'F-003',
        rule_id: 'complex-function',
        severity: 'high',
        category: 'quality',
        title: 'Function too complex',
        message: 'Function has cyclomatic complexity of 25',
        file: 'src/utils/helpers.js',
        line: 120,
        code_snippet: 'function processData(data) { // 150 lines... }'
      }
    ]
  }
}

onMounted(() => {
  loadScanData()
})
</script>
