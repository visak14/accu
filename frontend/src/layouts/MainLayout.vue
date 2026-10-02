<template>
  <q-layout view="hHh lpR fFf" class="main-layout bg-grey-2">
    <!-- Top Header -->
    <q-header elevated class="bg-slate-900 text-white header-bar">
      <q-toolbar class="q-px-lg">
        <!-- Logo & Title -->
        <div class="row items-center q-gutter-sm cursor-pointer" @click="activeTab = 'projects'">
          <q-avatar size="38px" color="primary" text-color="white" icon="science" />
          <div>
            <div class="text-subtitle1 text-weight-bolder tracking-wide">AccuScript SLR</div>
            <div class="text-caption text-primary-light text-weight-medium">AI Systematic Literature Screener</div>
          </div>
        </div>

        <q-space />

        <!-- Navigation Tabs -->
        <q-tabs
          v-model="activeTab"
          dense
          no-caps
          active-color="primary-light"
          indicator-color="primary"
          class="text-grey-4 gt-xs"
        >
          <q-tab name="projects" icon="folder_open" label="All Projects" />
          <q-tab name="upload" icon="cloud_upload" label="Upload Project" />
          <q-tab
            name="screening"
            icon="dashboard"
            label="Screening Dashboard"
            :disable="!currentProjectId"
          >
            <q-tooltip v-if="!currentProjectId">Select a project first to screen studies</q-tooltip>
          </q-tab>
        </q-tabs>

        <q-space />

        <!-- Right Side: Samples Menu & Status -->
        <div class="row items-center q-gutter-sm">
          <!-- Sample Files Download Dropdown -->
          <q-btn-dropdown
            flat
            dense
            no-caps
            icon="download"
            label="Sample Files"
            color="grey-4"
          >
            <q-list dense>
              <q-item clickable v-close-popup tag="a" :href="api.getSampleStudiesUrl()" target="_blank">
                <q-item-section avatar>
                  <q-icon name="table_view" color="primary" />
                </q-item-section>
                <q-item-section>
                  <q-item-label class="text-weight-bold">sample_studies.xlsx</q-item-label>
                  <q-item-label caption>12 candidate SLR studies dataset</q-item-label>
                </q-item-section>
              </q-item>
              <q-item clickable v-close-popup tag="a" :href="api.getSampleProtocolUrl()" target="_blank">
                <q-item-section avatar>
                  <q-icon name="rule" color="secondary" />
                </q-item-section>
                <q-item-section>
                  <q-item-label class="text-weight-bold">sample_protocol.xlsx</q-item-label>
                  <q-item-label caption>Review protocol & screening criteria</q-item-label>
                </q-item-section>
              </q-item>
            </q-list>
          </q-btn-dropdown>

          <!-- Backend Health Chip -->
          <q-badge
            :color="backendOnline ? 'positive' : 'negative'"
            class="q-pa-xs text-caption text-weight-bold"
          >
            <q-icon :name="backendOnline ? 'wifi' : 'wifi_off'" size="12px" class="q-mr-xs" />
            {{ backendOnline ? 'API Online' : 'Connecting...' }}
          </q-badge>
        </div>
      </q-toolbar>
    </q-header>

    <!-- Page Content Container -->
    <q-page-container>
      <!-- Upload Page -->
      <UploadProjectPage
        v-if="activeTab === 'upload'"
        @project-selected="onProjectSelected"
      />

      <!-- Screening Dashboard -->
      <ScreeningDashboardPage
        v-else-if="activeTab === 'screening' && currentProjectId"
        :key="currentProjectId"
        :project-id="currentProjectId"
      />

      <!-- Projects List Page (Default) -->
      <ProjectsListPage
        v-else
        @select-project="onProjectSelected"
        @create-project="activeTab = 'upload'"
      />
    </q-page-container>

    <!-- Footer -->
    <q-footer class="bg-white text-grey-7 border-top q-py-xs q-px-lg">
      <div class="row items-center justify-between text-caption">
        <div>
          <strong>SYMPRO Assignment</strong> • Literature Project Importer with Real LLM AI Screening
        </div>
        <div>
          Go (Chi) + MongoDB + Vue 3 (Quasar) + Multi-Provider LLMs
        </div>
      </div>
    </q-footer>
  </q-layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../services/api'
import UploadProjectPage from '../pages/UploadProjectPage.vue'
import ScreeningDashboardPage from '../pages/ScreeningDashboardPage.vue'
import ProjectsListPage from '../pages/ProjectsListPage.vue'

const activeTab = ref('projects')
const currentProjectId = ref(null)
const backendOnline = ref(false)

function onProjectSelected(projectId) {
  currentProjectId.value = projectId
  activeTab.value = 'screening'
}

async function checkHealth() {
  try {
    const res = await api.checkHealth()
    if (res.data?.status === 'ok') {
      backendOnline.value = true
    }
  } catch (err) {
    backendOnline.value = false
  }
}

onMounted(() => {
  checkHealth()
  setInterval(checkHealth, 10000)
})
</script>

<style scoped>
.bg-slate-900 {
  background-color: #0f172a !important;
}
.text-primary-light {
  color: #38bdf8;
}
.header-bar {
  border-bottom: 1px solid #1e293b;
}
.border-top {
  border-top: 1px solid #e2e8f0;
}
</style>
