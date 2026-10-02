<template>
  <div class="upload-page q-pa-md q-pa-lg-xl">
    <div class="row justify-center">
      <div class="col-12 col-md-10 col-lg-8">
        
        <!-- Header Banner -->
        <div class="q-mb-lg text-center">
          <div class="row items-center justify-center q-gutter-sm q-mb-xs">
            <q-avatar size="46px" color="primary" text-color="white" icon="cloud_upload" />
            <h1 class="text-h4 text-weight-bolder text-grey-9 q-my-none">Create Literature Review Project</h1>
          </div>
          <p class="text-subtitle1 text-grey-7">
            Upload your candidate studies Excel along with the SLR screening protocol to begin AI-assisted screening.
          </p>
        </div>

        <!-- Sample Download Quick-bar -->
        <q-banner rounded class="bg-blue-50 text-grey-9 q-mb-lg border-blue">
          <template v-slot:avatar>
            <q-icon name="download" color="primary" size="28px" />
          </template>
          <div class="text-weight-bold">Need sample test files?</div>
          <div class="text-caption text-grey-7 q-mb-xs">
            Download pre-configured SLR Excel templates (Type 2 Diabetes Telemedicine RCT dataset & review protocol):
          </div>
          <div class="row q-gutter-sm q-mt-xs">
            <q-btn
              outline
              dense
              no-caps
              size="sm"
              color="primary"
              icon="table_view"
              label="Download sample studies.xlsx"
              :href="api.getSampleStudiesUrl()"
              target="_blank"
            />
            <q-btn
              outline
              dense
              no-caps
              size="sm"
              color="secondary"
              icon="rule"
              label="Download sample protocol.xlsx"
              :href="api.getSampleProtocolUrl()"
              target="_blank"
            />
          </div>
        </q-banner>

        <!-- Main Upload Form Card -->
        <q-card flat bordered class="upload-card shadow-1">
          <q-card-section class="q-pa-lg">
            <q-form @submit.prevent="submitProject" class="q-gutter-y-lg">
              
              <!-- Section 1: Project Metadata -->
              <div>
                <div class="text-subtitle1 text-weight-bold text-primary q-mb-md row items-center q-gutter-xs">
                  <q-icon name="info" size="20px" />
                  <span>1. Project Details</span>
                </div>
                
                <div class="q-gutter-y-md">
                  <q-input
                    v-model="form.name"
                    outlined
                    label="Project Name *"
                    placeholder="e.g. Telemedicine Interventions for Type 2 Diabetes"
                    :rules="[val => !!val || 'Project name is required']"
                  >
                    <template v-slot:prepend>
                      <q-icon name="drive_file_rename_outline" />
                    </template>
                  </q-input>

                  <q-input
                    v-model="form.description"
                    outlined
                    type="textarea"
                    rows="3"
                    label="Project Description / Systematic Review Objective"
                    placeholder="A systematic literature review evaluating randomized controlled trials on digital health coaching vs standard clinical care..."
                  >
                    <template v-slot:prepend>
                      <q-icon name="description" />
                    </template>
                  </q-input>
                </div>
              </div>

              <q-separator />

              <!-- Section 2: LLM Configuration -->
              <div>
                <div class="text-subtitle1 text-weight-bold text-primary q-mb-md row items-center q-gutter-xs">
                  <q-icon name="psychology" size="20px" />
                  <span>2. Real LLM Screening Engine</span>
                </div>

                <div class="row q-col-gutter-md">
                  <div class="col-12 col-sm-6">
                    <q-select
                      v-model="form.llmProvider"
                      :options="providerOptions"
                      emit-value
                      map-options
                      outlined
                      label="AI Provider *"
                      @update:model-value="onProviderChange"
                    >
                      <template v-slot:prepend>
                        <q-icon name="smart_toy" />
                      </template>
                    </q-select>
                  </div>

                  <div class="col-12 col-sm-6">
                    <q-input
                      v-model="form.llmModel"
                      outlined
                      label="Model Name"
                      :placeholder="defaultModelForProvider(form.llmProvider)"
                    >
                      <template v-slot:prepend>
                        <q-icon name="tune" />
                      </template>
                    </q-input>
                  </div>

                  <div class="col-12">
                    <q-input
                      v-model="form.llmApiKey"
                      :type="showApiKey ? 'text' : 'password'"
                      outlined
                      label="LLM API Key"
                      placeholder="e.g. gsk_... or sk-..."
                      hint="Encrypted with AES-256 in MongoDB. If left empty, backend will use server .env fallback key if configured."
                    >
                      <template v-slot:prepend>
                        <q-icon name="key" />
                      </template>
                      <template v-slot:append>
                        <q-icon
                          :name="showApiKey ? 'visibility_off' : 'visibility'"
                          class="cursor-pointer"
                          @click="showApiKey = !showApiKey"
                        />
                      </template>
                    </q-input>
                  </div>
                </div>
              </div>

              <q-separator />

              <!-- Section 3: Dual Excel Upload -->
              <div>
                <div class="text-subtitle1 text-weight-bold text-primary q-mb-md row items-center q-gutter-xs">
                  <q-icon name="upload_file" size="20px" />
                  <span>3. Upload Excel Data Files (.xlsx)</span>
                </div>

                <div class="row q-col-gutter-md">
                  <!-- Studies Excel File -->
                  <div class="col-12 col-md-6">
                    <div class="file-drop-container q-pa-md rounded-borders text-center">
                      <q-file
                        v-model="form.studiesFile"
                        outlined
                        label="Studies Excel (.xlsx) *"
                        accept=".xlsx, .xls"
                        :rules="[val => !!val || 'Studies Excel file is required']"
                      >
                        <template v-slot:prepend>
                          <q-icon name="article" color="primary" />
                        </template>
                        <template v-slot:append>
                          <q-icon name="close" @click.stop.prevent="form.studiesFile = null" class="cursor-pointer" v-if="form.studiesFile" />
                        </template>
                      </q-file>
                      <div class="text-caption text-grey-7 q-mt-xs">
                        Required columns: <code>ID, title, year, author, abstract, article_type</code>
                      </div>
                    </div>
                  </div>

                  <!-- Protocol Excel File -->
                  <div class="col-12 col-md-6">
                    <div class="file-drop-container q-pa-md rounded-borders text-center">
                      <q-file
                        v-model="form.protocolFile"
                        outlined
                        label="Protocol Excel (.xlsx) *"
                        accept=".xlsx, .xls"
                        :rules="[val => !!val || 'Protocol Excel file is required']"
                      >
                        <template v-slot:prepend>
                          <q-icon name="fact_check" color="secondary" />
                        </template>
                        <template v-slot:append>
                          <q-icon name="close" @click.stop.prevent="form.protocolFile = null" class="cursor-pointer" v-if="form.protocolFile" />
                        </template>
                      </q-file>
                      <div class="text-caption text-grey-7 q-mt-xs">
                        Required columns: <code>criterion1, ..., inclusion_criteria, exclusion_criteria</code>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Upload Progress -->
              <div v-if="uploading" class="q-mt-md">
                <div class="row justify-between text-caption text-weight-bold text-primary q-mb-xs">
                  <span>Uploading and parsing SLR dataset...</span>
                  <span>{{ uploadProgress }}%</span>
                </div>
                <q-linear-progress :value="uploadProgress / 100" color="primary" stripe rounded size="8px" />
              </div>

              <!-- Submit Button -->
              <div class="row justify-end q-mt-lg">
                <q-btn
                  type="submit"
                  color="primary"
                  size="lg"
                  unelevated
                  icon="rocket_launch"
                  label="Import Project & Start Screening"
                  :loading="uploading"
                  class="full-width-sm"
                />
              </div>

            </q-form>
          </q-card-section>
        </q-card>
      </div>
    </div>

    <!-- Success Dialog -->
    <q-dialog v-model="successModal.open" persistent>
      <q-card style="min-width: 420px" class="rounded-borders text-center q-pa-md">
        <q-card-section>
          <q-avatar icon="check_circle" color="positive" text-color="white" size="64px" class="q-mb-md" />
          <div class="text-h5 text-weight-bold text-grey-9">Project Imported Successfully!</div>
          <div class="text-body1 text-grey-7 q-mt-sm">
            Imported <strong>{{ successModal.totalStudies }}</strong> candidate studies ready for SLR screening.
          </div>
          <div class="text-caption text-grey-6 q-mt-xs">
            Project ID: <code>{{ successModal.projectId }}</code>
          </div>
        </q-card-section>

        <q-card-actions align="center" class="q-gutter-sm">
          <q-btn
            unelevated
            color="primary"
            icon="dashboard"
            label="Open Screening Dashboard"
            @click="navigateToScreening(successModal.projectId)"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import api from '../services/api'
import { useQuasar } from 'quasar'

const $q = useQuasar()
const emit = defineEmits(['project-selected'])

const showApiKey = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)

const form = reactive({
  name: '',
  description: '',
  llmProvider: 'groq',
  llmModel: 'llama-3.3-70b-versatile',
  llmApiKey: '',
  studiesFile: null,
  protocolFile: null
})

const providerOptions = [
  { label: 'Groq Cloud (Fast & High Quality - Llama 3.3)', value: 'groq' },
  { label: 'OpenAI (GPT-4o-mini)', value: 'openai' },
  { label: 'Google Gemini (2.0 Flash / 1.5 Flash)', value: 'gemini' },
  { label: 'Anthropic Claude (3.5 Haiku)', value: 'claude' }
]

function defaultModelForProvider(prov) {
  switch (prov) {
    case 'groq': return 'llama-3.3-70b-versatile'
    case 'openai': return 'gpt-4o-mini'
    case 'gemini': return 'gemini-2.0-flash'
    case 'claude': return 'claude-3-5-haiku-20241022'
    default: return 'llama-3.3-70b-versatile'
  }
}

function onProviderChange(prov) {
  form.llmModel = defaultModelForProvider(prov)
}

const successModal = reactive({
  open: false,
  projectId: '',
  totalStudies: 0
})

async function submitProject() {
  if (!form.studiesFile || !form.protocolFile) {
    $q.notify({
      type: 'warning',
      message: 'Please select both Studies and Protocol Excel files'
    })
    return
  }

  uploading.value = true
  uploadProgress.value = 0

  try {
    const data = new FormData()
    data.append('name', form.name)
    data.append('description', form.description)
    data.append('llmProvider', form.llmProvider)
    data.append('llmModel', form.llmModel)
    data.append('llmApiKey', form.llmApiKey)
    data.append('studies', form.studiesFile)
    data.append('protocol', form.protocolFile)

    const res = await api.uploadProject(data, (percent) => {
      uploadProgress.value = percent
    })

    successModal.projectId = res.data.projectId
    successModal.totalStudies = res.data.totalStudies
    successModal.open = true

    $q.notify({
      type: 'positive',
      message: `Project created with ${res.data.totalStudies} studies!`
    })
  } catch (err) {
    $q.notify({
      type: 'negative',
      message: err.response?.data?.error || err.message || 'Upload failed'
    })
  } finally {
    uploading.value = false
  }
}

function navigateToScreening(projectId) {
  successModal.open = false
  emit('project-selected', projectId)
}
</script>

<style scoped>
.upload-card {
  border-radius: 12px;
  background: white;
}
.border-blue {
  border: 1px solid #bfdbfe;
  background-color: #f0f9ff;
}
.file-drop-container {
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
}
code {
  background: #e2e8f0;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.85em;
}
</style>
