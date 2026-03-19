import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useScanStore = defineStore('scan', () => {
  // State
  const scanResult = ref(null)
  const loading = ref(false)
  const error = ref(null)
  const selectedFinding = ref(null)

  // Getters
  const findings = computed(() => scanResult.value?.findings || [])
  const healthScore = computed(() => scanResult.value?.health_score || {
    overall: 0,
    security: 0,
    quality: 0,
    architecture: 0,
    performance: 0
  })
  const summary = computed(() => scanResult.value?.summary || {
    critical: 0,
    high: 0,
    medium: 0,
    low: 0,
    info: 0,
    total: 0
  })
  const project = computed(() => scanResult.value?.project || {
    name: '',
    path: '',
    languages: [],
    files_scanned: 0
  })

  // Actions
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
    } finally {
      loading.value = false
    }
  }

  const refreshScan = () => {
    loadScanData()
  }

  const selectFinding = (finding) => {
    selectedFinding.value = finding
  }

  const clearSelectedFinding = () => {
    selectedFinding.value = null
  }

  return {
    scanResult,
    loading,
    error,
    selectedFinding,
    findings,
    healthScore,
    summary,
    project,
    loadScanData,
    refreshScan,
    selectFinding,
    clearSelectedFinding
  }
})
