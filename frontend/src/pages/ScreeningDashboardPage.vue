<template>
  <div class="screening-dashboard q-pa-md">
    <!-- Top Project Bar -->
    <div class="project-header q-pa-md bg-white rounded-borders shadow-1 q-mb-md">
      <div class="row items-center justify-between wrap q-gutter-y-sm">
        <div>
          <div class="row items-center q-gutter-xs">
            <q-badge color="primary" class="q-mr-xs text-weight-bold">SLR PROJECT</q-badge>
            <h2 class="text-h6 text-weight-bolder text-grey-9 q-my-none">{{ project?.name || 'Loading Project...' }}</h2>
          </div>
          <div class="text-caption text-grey-7 q-mt-xs">{{ project?.description || 'Systematic Literature Review Screening' }}</div>
        </div>

        <div class="row items-center q-gutter-sm">
          <!-- Batch Screening Button -->
          <q-btn
            unelevated
            color="deep-purple-7"
            icon="auto_awesome_motion"
            label="Batch AI Screen"
            no-caps
            :disable="!project || undecidedCount === 0"
            @click="showBatchDialog = true"
          >
            <q-tooltip>Screen multiple undecided studies automatically</q-tooltip>
          </q-btn>

          <!-- Export to Excel Button -->
          <q-btn
            unelevated
            color="positive"
            icon="file_download"
            label="Export Excel"
            no-caps
            :href="api.getExportUrl(projectId)"
            target="_blank"
          >
            <q-tooltip>Download decisions and AI audit log to styled Excel (.xlsx)</q-tooltip>
          </q-btn>
        </div>
      </div>

      <!-- Stats Bar & Progress -->
      <div class="row items-center justify-between q-mt-md q-pt-sm border-top">
        <div class="row items-center q-gutter-md">
          <div class="stat-badge">
            <span class="text-caption text-grey-7">Total: </span>
            <span class="text-weight-bold text-grey-9">{{ totalStudies }}</span>
          </div>
          <div class="stat-badge">
            <q-badge rounded color="positive" class="q-mr-xs" />
            <span class="text-caption text-grey-7">Included: </span>
            <span class="text-weight-bolder text-positive">{{ includedCount }}</span>
          </div>
          <div class="stat-badge">
            <q-badge rounded color="negative" class="q-mr-xs" />
            <span class="text-caption text-grey-7">Excluded: </span>
            <span class="text-weight-bolder text-negative">{{ excludedCount }}</span>
          </div>
          <div class="stat-badge">
            <q-badge rounded color="grey-6" class="q-mr-xs" />
            <span class="text-caption text-grey-7">Undecided: </span>
            <span class="text-weight-bolder text-grey-8">{{ undecidedCount }}</span>
          </div>
          <div class="stat-badge">
            <q-badge rounded color="accent" class="q-mr-xs" />
            <span class="text-caption text-grey-7">AI Screened: </span>
            <span class="text-weight-bolder text-accent">{{ aiScreenedCount }}</span>
          </div>
        </div>

        <div class="row items-center q-gutter-sm" style="min-width: 220px">
          <span class="text-caption text-weight-bold text-grey-7">{{ progressPercent }}% Screened</span>
          <q-linear-progress
            :value="progressPercent / 100"
            color="positive"
            track-color="grey-3"
            rounded
            size="8px"
            style="width: 120px"
          />
        </div>
      </div>
    </div>

    <!-- Main Dual Panel Split View (60% / 40%) -->
    <div class="row q-col-gutter-md">
      
      <!-- LEFT PANEL (60%): Studies Table Grid -->
      <div class="col-12 col-lg-7">
        <q-card flat bordered class="studies-table-card full-height">
          <!-- Filter Tabs & Search Bar -->
          <div class="q-pa-sm bg-grey-1 border-bottom">
            <div class="row items-center justify-between q-col-gutter-sm">
              <div class="col-12 col-sm-7">
                <q-tabs
                  v-model="decisionFilter"
                  dense
                  no-caps
                  active-color="primary"
                  indicator-color="primary"
                  class="text-grey-7"
                  @update:model-value="fetchStudies"
                >
                  <q-tab name="all" :label="`All (${totalStudies})`" />
                  <q-tab name="undecided" :label="`Undecided (${undecidedCount})`" />
                  <q-tab name="included" :label="`Included (${includedCount})`" />
                  <q-tab name="excluded" :label="`Excluded (${excludedCount})`" />
                </q-tabs>
              </div>

              <div class="col-12 col-sm-5">
                <q-input
                  v-model="searchQuery"
                  dense
                  outlined
                  placeholder="Search title, author, abstract..."
                  @update:model-value="debouncedSearch"
                  clearable
                >
                  <template v-slot:prepend>
                    <q-icon name="search" size="18px" />
                  </template>
                </q-input>
              </div>
            </div>
          </div>

          <!-- QTable -->
          <q-table
            flat
            :rows="studies"
            :columns="columns"
            row-key="studyId"
            :loading="loadingTable"
            v-model:pagination="pagination"
            @request="onTableRequest"
            binary-state-sort
            :rows-per-page-options="[10, 25, 50, 100]"
            class="studies-table"
          >
            <!-- Custom Row Click & Selection Highlight -->
            <template v-slot:body="props">
              <q-tr
                :props="props"
                @click="selectStudy(props.row)"
                class="cursor-pointer study-row"
                :class="{ 'selected-row': selectedStudy?.studyId === props.row.studyId }"
              >
                <q-td key="ID" :props="props" class="text-weight-medium">
                  #{{ props.row.ID || props.row.RawID }}
                </q-td>

                <q-td key="title" :props="props" class="title-cell">
                  <div class="text-weight-bold text-grey-9 line-clamp-2">
                    {{ props.row.title }}
                  </div>
                  <div class="text-caption text-grey-7 q-mt-xs">
                    {{ props.row.author }} • {{ props.row.year }} • <span class="text-primary">{{ props.row.article_type }}</span>
                  </div>
                </q-td>

                <q-td key="decision" :props="props">
                  <q-badge
                    :color="decisionBadgeColor(props.row.decision)"
                    class="q-pa-xs text-weight-bold text-uppercase"
                  >
                    {{ props.row.decision }}
                  </q-badge>
                </q-td>

                <q-td key="aiSuggestion" :props="props">
                  <div v-if="props.row.aiSuggestion" class="row items-center q-gutter-xs no-wrap">
                    <q-badge
                      :color="props.row.aiSuggestion === 'include' ? 'positive' : 'negative'"
                      class="text-weight-bold"
                    >
                      <q-icon
                        :name="props.row.aiSuggestion === 'include' ? 'check' : 'close'"
                        size="14px"
                        class="q-mr-xs"
                      />
                      {{ props.row.aiSuggestion.toUpperCase() }}
                    </q-badge>
                    <span class="text-caption text-weight-bold text-grey-8" v-if="props.row.aiConfidence">
                      {{ Math.round(props.row.aiConfidence * 100) }}%
                    </span>
                  </div>
                  <span v-else class="text-caption text-grey-5">—</span>
                </q-td>
              </q-tr>
            </template>
          </q-table>
        </q-card>
      </div>

      <!-- RIGHT PANEL (40%): Active Study Viewer & Screening Actions -->
      <div class="col-12 col-lg-5">
        <div v-if="selectedStudy" class="column q-gutter-y-md">
          
          <!-- Study Detail Card -->
          <q-card flat bordered class="active-study-card bg-white">
            <q-card-section class="q-pb-sm">
              <div class="row items-center justify-between q-mb-xs">
                <div class="row items-center q-gutter-xs">
                  <q-badge color="blue-grey-8" class="text-weight-bold">
                    STUDY #{{ selectedStudy.ID || selectedStudy.RawID }}
                  </q-badge>
                  <q-badge color="blue-1" text-color="primary" class="text-weight-medium">
                    {{ selectedStudy.article_type }}
                  </q-badge>
                  <q-badge color="grey-2" text-color="grey-9">
                    {{ selectedStudy.year }}
                  </q-badge>
                </div>

                <!-- Nav buttons -->
                <div class="row items-center q-gutter-xs">
                  <q-btn
                    flat
                    dense
                    round
                    icon="chevron_left"
                    :disable="currentIndex <= 0"
                    @click="goToPrevStudy"
                  >
                    <q-tooltip>Previous Study (Left Arrow)</q-tooltip>
                  </q-btn>
                  <span class="text-caption text-grey-7">{{ currentIndex + 1 }} / {{ studies.length }}</span>
                  <q-btn
                    flat
                    dense
                    round
                    icon="chevron_right"
                    :disable="currentIndex >= studies.length - 1"
                    @click="goToNextStudy"
                  >
                    <q-tooltip>Next Study (Right Arrow)</q-tooltip>
                  </q-btn>
                </div>
              </div>

              <h3 class="text-subtitle1 text-weight-bolder text-grey-9 q-mt-xs q-mb-xs leading-normal">
                {{ selectedStudy.title }}
              </h3>
              <div class="text-caption text-weight-medium text-grey-8">
                <q-icon name="person" size="14px" /> {{ selectedStudy.author }} ({{ selectedStudy.year }})
              </div>
            </q-card-section>

            <q-separator />

            <q-card-section class="q-pt-sm">
              <div class="text-caption text-weight-bold text-uppercase text-grey-7 q-mb-xs">Abstract</div>
              <div class="abstract-text text-body2 text-grey-9">
                {{ selectedStudy.abstract || 'No abstract text available for this study.' }}
              </div>
            </q-card-section>
          </q-card>

          <!-- Protocol Criteria Reference Accordion -->
          <ProtocolViewer :protocol="project?.protocol" />

          <!-- AI Screening Suggestion Card -->
          <AISuggestionCard
            :study="selectedStudy"
            :loading="aiLoading"
            @request-ai-suggest="runAISuggestion"
          />

          <!-- Action Bar: Researcher Screening Decision -->
          <q-card flat bordered class="decision-action-card q-pa-md bg-white">
            <div class="text-caption text-weight-bold text-uppercase text-grey-8 q-mb-sm text-center">
              Researcher Decision (Hotkeys: [I] Include, [E] Exclude)
            </div>

            <div class="row q-gutter-sm justify-center">
              <q-btn
                unelevated
                size="md"
                color="positive"
                icon="check"
                label="Include (I)"
                :loading="updatingDecision"
                :outline="selectedStudy.decision !== 'included'"
                class="decision-btn"
                @click="setDecision('included')"
              />

              <q-btn
                unelevated
                size="md"
                color="negative"
                icon="close"
                label="Exclude (E)"
                :loading="updatingDecision"
                :outline="selectedStudy.decision !== 'excluded'"
                class="decision-btn"
                @click="setDecision('excluded')"
              />

              <q-btn
                flat
                dense
                color="grey-7"
                icon="restart_alt"
                label="Reset"
                :disable="selectedStudy.decision === 'undecided'"
                @click="setDecision('undecided')"
              >
                <q-tooltip>Reset decision to undecided</q-tooltip>
              </q-btn>
            </div>
          </q-card>

        </div>

        <!-- No Study Selected State -->
        <div v-else class="column items-center justify-center q-pa-xl text-center text-grey-6 bg-white rounded-borders border">
          <q-icon name="touch_app" size="48px" class="q-mb-md" />
          <div class="text-h6 text-weight-medium">Select a Study to Begin Screening</div>
          <div class="text-caption text-grey-7">Click any row in the left studies table to view details and AI suggestions</div>
        </div>

      </div>
    </div>

    <!-- Batch Screening Dialog -->
    <BatchScreeningDialog
      v-model="showBatchDialog"
      :project-id="projectId"
      :undecided-count="undecidedCount"
      @batch-finished="onBatchFinished"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import api from '../services/api'
import { useQuasar } from 'quasar'
import AISuggestionCard from '../components/AISuggestionCard.vue'
import ProtocolViewer from '../components/ProtocolViewer.vue'
import BatchScreeningDialog from '../components/BatchScreeningDialog.vue'

const $q = useQuasar()

const props = defineProps({
  projectId: {
    type: String,
    required: true
  }
})

const project = ref(null)
const studies = ref([])
const selectedStudy = ref(null)
const loadingTable = ref(false)
const aiLoading = ref(false)
const updatingDecision = ref(false)
const showBatchDialog = ref(false)

const decisionFilter = ref('all')
const searchQuery = ref('')
let searchTimeout = null

const pagination = ref({
  page: 1,
  rowsPerPage: 25,
  rowsNumber: 0
})

const columns = [
  { name: 'ID', label: '#', field: 'ID', align: 'left', sortable: true, style: 'width: 60px' },
  { name: 'title', label: 'Title & Author', field: 'title', align: 'left', sortable: true },
  { name: 'decision', label: 'Decision', field: 'decision', align: 'center', sortable: true, style: 'width: 110px' },
  { name: 'aiSuggestion', label: 'AI Suggestion', field: 'aiSuggestion', align: 'center', sortable: true, style: 'width: 140px' }
]

// Stats
const totalStudies = computed(() => project.value?.totalStudies || studies.value.length)
const includedCount = computed(() => project.value?.includedCount || 0)
const excludedCount = computed(() => project.value?.excludedCount || 0)
const undecidedCount = computed(() => project.value?.undecidedCount || 0)
const aiScreenedCount = computed(() => project.value?.aiScreenedCount || 0)

const progressPercent = computed(() => {
  if (!totalStudies.value) return 0
  const screened = includedCount.value + excludedCount.value
  return Math.round((screened / totalStudies.value) * 100)
})

const currentIndex = computed(() => {
  if (!selectedStudy.value) return -1
  return studies.value.findIndex(s => s.studyId === selectedStudy.value.studyId)
})

function decisionBadgeColor(decision) {
  switch (decision?.toLowerCase()) {
    case 'included': return 'positive'
    case 'excluded': return 'negative'
    default: return 'grey-6'
  }
}

async function fetchProject() {
  try {
    const res = await api.getProject(props.projectId)
    project.value = res.data
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Failed to load project details' })
  }
}

async function fetchStudies() {
  loadingTable.value = true
  try {
    const params = {
      decision: decisionFilter.value,
      search: searchQuery.value,
      page: pagination.value.page,
      limit: pagination.value.rowsPerPage
    }
    const res = await api.getStudies(props.projectId, params)
    studies.value = res.data.data || []
    pagination.value.rowsNumber = res.data.total || 0

    // Auto-select first study if none selected
    if (studies.value.length > 0 && (!selectedStudy.value || !studies.value.some(s => s.studyId === selectedStudy.value.studyId))) {
      selectedStudy.value = studies.value[0]
    }
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Failed to load studies list' })
  } finally {
    loadingTable.value = false
  }
}

function onTableRequest(props) {
  pagination.value = props.pagination
  fetchStudies()
}

function debouncedSearch() {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    fetchStudies()
  }, 350)
}

function selectStudy(study) {
  selectedStudy.value = study
}

function goToPrevStudy() {
  const idx = currentIndex.value
  if (idx > 0) {
    selectedStudy.value = studies.value[idx - 1]
  }
}

function goToNextStudy() {
  const idx = currentIndex.value
  if (idx >= 0 && idx < studies.value.length - 1) {
    selectedStudy.value = studies.value[idx + 1]
  }
}

async function setDecision(decision) {
  if (!selectedStudy.value) return
  updatingDecision.value = true

  try {
    const res = await api.updateDecision(selectedStudy.value.studyId, decision)
    
    // Update local state
    selectedStudy.value.decision = decision
    selectedStudy.value.decidedAt = res.data.decidedAt
    
    const studyInList = studies.value.find(s => s.studyId === selectedStudy.value.studyId)
    if (studyInList) {
      studyInList.decision = decision
    }

    await fetchProject()

    $q.notify({
      type: decision === 'included' ? 'positive' : decision === 'excluded' ? 'negative' : 'info',
      message: `Marked study as ${decision.toUpperCase()}`,
      timeout: 1200
    })

    // Advance to next study automatically on decision
    if (decision !== 'undecided') {
      goToNextStudy()
    }
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Failed to update decision' })
  } finally {
    updatingDecision.value = false
  }
}

async function runAISuggestion(overrides = {}) {
  if (!selectedStudy.value) return
  aiLoading.value = true

  try {
    const payload = {}
    if (overrides.provider) payload.provider = overrides.provider
    if (overrides.apiKey) payload.apiKey = overrides.apiKey

    const res = await api.getAISuggestion(selectedStudy.value.studyId, payload)
    
    // Update active study
    selectedStudy.value.aiSuggestion = res.data.suggestion
    selectedStudy.value.aiConfidence = res.data.confidence
    selectedStudy.value.aiReason = res.data.aiReason
    selectedStudy.value.aiMatches = res.data.matches
    selectedStudy.value.aiJsonResponse = res.data.aiJsonResponse

    // Update in table list
    const studyInList = studies.value.find(s => s.studyId === selectedStudy.value.studyId)
    if (studyInList) {
      studyInList.aiSuggestion = res.data.suggestion
      studyInList.aiConfidence = res.data.confidence
      studyInList.aiReason = res.data.aiReason
      studyInList.aiMatches = res.data.matches
      studyInList.aiJsonResponse = res.data.aiJsonResponse
    }

    await fetchProject()

    $q.notify({
      type: 'positive',
      message: `AI Suggestion: ${res.data.suggestion.toUpperCase()} (${Math.round(res.data.confidence * 100)}% confidence)`
    })
  } catch (err) {
    $q.notify({
      type: 'negative',
      message: err.response?.data?.error || 'AI screening request failed'
    })
  } finally {
    aiLoading.value = false
  }
}

function onBatchFinished() {
  fetchProject()
  fetchStudies()
}

// Global Keyboard Shortcuts
function handleKeydown(e) {
  // Avoid capturing hotkeys when typing in input or textarea
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName)) {
    return
  }

  if (e.key === 'i' || e.key === 'I') {
    e.preventDefault()
    setDecision('included')
  } else if (e.key === 'e' || e.key === 'E') {
    e.preventDefault()
    setDecision('excluded')
  } else if (e.key === 'a' || e.key === 'A') {
    e.preventDefault()
    runAISuggestion()
  } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
    e.preventDefault()
    goToPrevStudy()
  } else if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
    e.preventDefault()
    goToNextStudy()
  }
}

onMounted(() => {
  fetchProject()
  fetchStudies()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.border-top {
  border-top: 1px solid #f1f5f9;
}
.border-bottom {
  border-bottom: 1px solid #e2e8f0;
}
.studies-table-card {
  border-radius: 10px;
  background: white;
}
.studies-table :deep(th) {
  font-weight: 700;
  color: #475569;
  background-color: #f8fafc;
}
.study-row:hover {
  background-color: #f0f9ff !important;
}
.selected-row {
  background-color: #e0f2fe !important;
  border-left: 4px solid #0284c7;
}
.title-cell {
  max-width: 320px;
  white-space: normal;
}
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.active-study-card {
  border-radius: 8px;
}
.abstract-text {
  line-height: 1.6;
  max-height: 250px;
  overflow-y: auto;
}
.decision-action-card {
  border-radius: 8px;
}
.decision-btn {
  min-width: 140px;
}
</style>
