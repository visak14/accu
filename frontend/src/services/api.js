import axios from 'axios'

const apiClient = axios.create({
  baseURL: '/api',
  timeout: 60000,
  headers: {
    'Accept': 'application/json'
  }
})

export default {
  // Health
  checkHealth() {
    return apiClient.get('/health')
  },

  // Projects
  uploadProject(formData, onProgress) {
    return apiClient.post('/projects/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(percent)
        }
      }
    })
  },

  getProjects() {
    return apiClient.get('/projects')
  },

  getProject(id) {
    return apiClient.get(`/projects/${id}`)
  },

  deleteProject(id) {
    return apiClient.delete(`/projects/${id}`)
  },

  // Studies
  getStudies(projectId, params = {}) {
    return apiClient.get(`/projects/${projectId}/studies`, { params })
  },

  getStudy(studyId) {
    return apiClient.get(`/studies/${studyId}`)
  },

  updateDecision(studyId, decision) {
    return apiClient.patch(`/studies/${studyId}/decision`, { decision })
  },

  // AI Screening
  getAISuggestion(studyId, data = {}) {
    return apiClient.post(`/studies/${studyId}/ai-suggest`, data)
  },

  batchAISuggest(projectId, options = {}) {
    return apiClient.post(`/projects/${projectId}/batch-ai-suggest`, options)
  },

  // Export URL helper
  getExportUrl(projectId, filter = 'all') {
    return `/api/projects/${projectId}/export?filter=${filter}`
  },

  // Sample files download helpers
  getSampleStudiesUrl() {
    return '/api/sample-files/studies'
  },

  getSampleProtocolUrl() {
    return '/api/sample-files/protocol'
  }
}
