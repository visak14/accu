<template>
  <div class="projects-page q-pa-md q-pa-lg-xl">
    <div class="row items-center justify-between q-mb-lg">
      <div>
        <h1 class="text-h4 text-weight-bolder text-grey-9 q-my-none">Literature Review Projects</h1>
        <div class="text-subtitle1 text-grey-7">Manage your systematic literature reviews, check screening progress, and export results</div>
      </div>
      <q-btn
        unelevated
        color="primary"
        icon="add"
        label="New Review Project"
        size="md"
        @click="$emit('create-project')"
      />
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="column items-center justify-center q-py-xl text-center">
      <q-spinner-dots color="primary" size="48px" />
      <div class="text-body1 text-grey-7 q-mt-md">Loading projects...</div>
    </div>

    <!-- Empty State -->
    <div v-else-if="projects.length === 0" class="column items-center justify-center q-py-xl text-center bg-white rounded-borders border q-pa-xl shadow-1">
      <q-avatar icon="library_books" color="blue-1" text-color="primary" size="72px" class="q-mb-md" />
      <div class="text-h5 text-weight-bold text-grey-9">No Review Projects Yet</div>
      <div class="text-body1 text-grey-7 q-mt-xs q-mb-lg max-w-md">
        Upload your candidate studies Excel file and systematic review protocol to begin AI-powered article screening.
      </div>
      <q-btn
        unelevated
        color="primary"
        icon="cloud_upload"
        label="Upload Your First Project"
        size="lg"
        @click="$emit('create-project')"
      />
    </div>

    <!-- Projects Grid -->
    <div v-else class="row q-col-gutter-lg">
      <div v-for="proj in projects" :key="proj.projectId" class="col-12 col-md-6 col-lg-4">
        <q-card flat bordered class="project-card column justify-between shadow-1">
          <q-card-section>
            <!-- Top Tag & Date -->
            <div class="row items-center justify-between q-mb-sm">
              <q-badge color="primary" class="text-weight-bold">
                {{ proj.llmProvider ? proj.llmProvider.toUpperCase() : 'GROQ' }}
              </q-badge>
              <div class="text-caption text-grey-6">
                {{ formatDate(proj.createdAt) }}
              </div>
            </div>

            <!-- Project Title & Description -->
            <h2 class="text-h6 text-weight-bolder text-grey-9 q-mb-xs title-text">
              {{ proj.name }}
            </h2>
            <p class="text-caption text-grey-7 description-text q-mb-md">
              {{ proj.description || 'No description provided' }}
            </p>

            <!-- Screening Progress -->
            <div class="q-mb-md">
              <div class="row justify-between text-caption text-weight-bold text-grey-8 q-mb-xs">
                <span>Screening Progress</span>
                <span>{{ calculateProgress(proj) }}%</span>
              </div>
              <q-linear-progress
                :value="calculateProgress(proj) / 100"
                color="positive"
                track-color="grey-2"
                rounded
                size="8px"
              />
            </div>

            <!-- Metrics Pills -->
            <div class="row q-gutter-xs wrap">
              <q-badge color="grey-2" text-color="grey-9" class="q-pa-xs">
                Total: <strong>{{ proj.totalStudies }}</strong>
              </q-badge>
              <q-badge color="green-1" text-color="green-9" class="q-pa-xs">
                Included: <strong>{{ proj.includedCount || 0 }}</strong>
              </q-badge>
              <q-badge color="red-1" text-color="red-9" class="q-pa-xs">
                Excluded: <strong>{{ proj.excludedCount || 0 }}</strong>
              </q-badge>
              <q-badge color="blue-1" text-color="blue-9" class="q-pa-xs">
                Undecided: <strong>{{ proj.undecidedCount || 0 }}</strong>
              </q-badge>
            </div>
          </q-card-section>

          <q-separator />

          <!-- Card Actions -->
          <q-card-actions class="q-pa-md row items-center justify-between bg-grey-1">
            <q-btn
              unelevated
              color="primary"
              icon="dashboard"
              label="Screen Studies"
              no-caps
              size="sm"
              @click="$emit('select-project', proj.projectId)"
            />

            <div class="row items-center q-gutter-xs">
              <!-- Export Excel -->
              <q-btn
                flat
                round
                dense
                color="positive"
                icon="file_download"
                :href="api.getExportUrl(proj.projectId)"
                target="_blank"
              >
                <q-tooltip>Download Decisions (.xlsx)</q-tooltip>
              </q-btn>

              <!-- Delete Project -->
              <q-btn
                flat
                round
                dense
                color="negative"
                icon="delete_outline"
                @click="confirmDelete(proj)"
              >
                <q-tooltip>Delete Project</q-tooltip>
              </q-btn>
            </div>
          </q-card-actions>
        </q-card>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../services/api'
import { useQuasar } from 'quasar'

const $q = useQuasar()
const emit = defineEmits(['select-project', 'create-project'])

const projects = ref([])
const loading = ref(false)

async function fetchProjects() {
  loading.value = true
  try {
    const res = await api.getProjects()
    projects.value = res.data || []
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Failed to load projects' })
  } finally {
    loading.value = false
  }
}

function calculateProgress(proj) {
  if (!proj.totalStudies) return 0
  const screened = (proj.includedCount || 0) + (proj.excludedCount || 0)
  return Math.round((screened / proj.totalStudies) * 100)
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function confirmDelete(proj) {
  $q.dialog({
    title: 'Confirm Delete',
    message: `Are you sure you want to permanently delete project "${proj.name}" and all its ${proj.totalStudies} studies?`,
    cancel: true,
    persistent: true,
    ok: {
      color: 'negative',
      label: 'Delete',
      unelevated: true
    }
  }).onOk(async () => {
    try {
      await api.deleteProject(proj.projectId)
      $q.notify({ type: 'positive', message: 'Project deleted successfully' })
      fetchProjects()
    } catch (err) {
      $q.notify({ type: 'negative', message: 'Failed to delete project' })
    }
  })
}

onMounted(() => {
  fetchProjects()
})
</script>

<style scoped>
.project-card {
  border-radius: 12px;
  background: white;
  min-height: 280px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.project-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}
.title-text {
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.description-text {
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
