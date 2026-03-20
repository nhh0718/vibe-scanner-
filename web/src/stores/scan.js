import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useScanStore = defineStore('scan', () => {
  // State
  const scanResult = ref(null)
  const reports = ref([])
  const activeReportId = ref(null)
  const loading = ref(false)
  const error = ref(null)
  const selectedFinding = ref(null)

  // Getters
  const findings = computed(() => scanResult.value?.findings || [])
  const healthScore = computed(() => scanResult.value?.health_score || {
    overall: 0, security: 0, quality: 0, architecture: 0, performance: 0
  })
  const summary = computed(() => scanResult.value?.summary || {
    critical: 0, high: 0, medium: 0, low: 0, info: 0, total: 0
  })
  const project = computed(() => scanResult.value?.project || {
    name: '', path: '', languages: [], files_scanned: 0
  })

  // Actions
  const loadScanData = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/scan')
      if (!response.ok) throw new Error('Không thể tải dữ liệu scan')
      scanResult.value = await response.json()
      activeReportId.value = scanResult.value?.scan_id || null
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  const loadReports = async () => {
    try {
      const response = await fetch('/api/reports')
      if (response.ok) {
        reports.value = await response.json()
      }
    } catch (err) {
      console.error('Failed to load reports:', err)
    }
  }

  const switchReport = async (filename) => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`/api/reports/${filename}/activate`, { method: 'POST' })
      if (!response.ok) throw new Error('Không thể chuyển báo cáo')
      // Reload current scan data after switching
      await loadScanData()
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  const refreshToLatest = async () => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/refresh', { method: 'POST' })
      if (!response.ok) throw new Error('Không thể làm mới dữ liệu')
      scanResult.value = await response.json()
      activeReportId.value = scanResult.value?.scan_id || null
      await loadReports()
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  const deleteReport = async (identifier) => {
    try {
      const response = await fetch(`/api/reports/${identifier}`, { method: 'DELETE' })
      if (response.ok) {
        await loadReports()
      }
    } catch (err) {
      console.error('Failed to delete report:', err)
    }
  }

  const selectFinding = (finding) => { selectedFinding.value = finding }
  const clearSelectedFinding = () => { selectedFinding.value = null }

  return {
    scanResult, reports, activeReportId, loading, error, selectedFinding,
    findings, healthScore, summary, project,
    loadScanData, loadReports, switchReport, refreshToLatest, deleteReport,
    selectFinding, clearSelectedFinding
  }
})
