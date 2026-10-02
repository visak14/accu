<template>
  <q-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)" persistent>
    <q-card style="min-width: 500px; max-width: 600px" class="rounded-borders">
      <q-card-section class="row items-center q-pb-none">
        <div class="row items-center q-gutter-sm">
          <q-avatar icon="auto_mode" color="primary" text-color="white" size="36px" />
          <div>
            <div class="text-h6 text-weight-bold">Batch AI Screening</div>
            <div class="text-caption text-grey-7">Screen multiple undecided studies in sequence</div>
          </div>
        </div>
        <q-space />
        <q-btn icon="close" flat round dense v-close-popup :disable="isRunning" />
      </q-card-section>

      <q-card-section class="q-pt-md">
        <!-- Settings Form -->
        <div v-if="!isRunning && !completed" class="column q-gutter-y-md">
          <q-banner dense class="bg-blue-1 text-primary rounded-borders">
            <template v-slot:avatar>
              <q-icon name="info" color="primary" />
            </template>
            The assistant will evaluate each study against the project protocol and record recommendations with confidence scores.
          </q-banner>

          <div>
            <div class="text-caption text-weight-bold text-grey-8 q-mb-xs">Studies to Screen</div>
            <q-slider
              v-model="batchLimit"
              :min="1"
              :max="Math.min(undecidedCount || 10, 30)"
              :step="1"
              label
              label-always
              color="primary"
            />
            <div class="text-caption text-grey-6 text-right">
              {{ batchLimit }} of {{ undecidedCount }} undecided studies
            </div>
          </div>

          <q-toggle
            v-model="unscreenedOnly"
            label="Only screen studies that have no AI suggestions yet"
            color="primary"
          />

          <!-- LLM Provider Override -->
          <div class="row q-col-gutter-sm">
            <div class="col-12 col-sm-6">
              <q-select
                v-model="selectedProvider"
                :options="providerOptions"
                emit-value
                map-options
                dense
                outlined
                label="AI Provider"
              />
            </div>
            <div class="col-12 col-sm-6">
              <q-input
                v-model="customApiKey"
                type="password"
                dense
                outlined
                label="API Key (optional override)"
                placeholder="Leave blank to use project key"
              />
            </div>
          </div>
        </div>

        <!-- Progress View -->
        <div v-if="isRunning" class="column items-center justify-center q-py-lg text-center">
          <q-spinner-dots color="primary" size="48px" />
          <div class="text-subtitle1 text-weight-bold q-mt-md text-primary">
            Screening Studies with Real LLM...
          </div>
          <div class="text-caption text-grey-7 q-mt-xs">
            Evaluating abstract semantics, protocol criteria, and reasoning
          </div>
        </div>

        <!-- Completion View -->
        <div v-if="completed" class="column q-gutter-y-sm">
          <div class="row items-center q-gutter-sm text-positive q-mb-sm">
            <q-icon name="task_alt" size="28px" />
            <span class="text-subtitle1 text-weight-bold">Batch Screening Completed!</span>
          </div>
          <div class="text-body2">
            Successfully evaluated <strong>{{ batchResults.length }}</strong> studies.
          </div>
          <q-scroll-area style="height: 180px;" class="bg-grey-1 rounded-borders q-pa-xs">
            <q-list dense separator>
              <q-item v-for="(item, idx) in batchResults" :key="idx">
                <q-item-section avatar>
                  <q-icon
                    :name="item.suggestion === 'include' ? 'check_circle' : 'cancel'"
                    :color="item.suggestion === 'include' ? 'positive' : 'negative'"
                    size="20px"
                  />
                </q-item-section>
                <q-item-section>
                  <q-item-label class="text-caption text-weight-medium ellipsis">{{ item.title }}</q-item-label>
                  <q-item-label caption class="ellipsis">{{ item.reasoning || item.error }}</q-item-label>
                </q-item-section>
                <q-item-section side>
                  <q-badge
                    :color="item.suggestion === 'include' ? 'positive' : 'negative'"
                    class="text-weight-bold"
                  >
                    {{ item.suggestion ? item.suggestion.toUpperCase() : 'ERROR' }}
                  </q-badge>
                </q-item-section>
              </q-item>
            </q-list>
          </q-scroll-area>
        </div>
      </q-card-section>

      <q-separator />

      <q-card-actions align="right" class="q-pa-md">
        <q-btn
          flat
          label="Cancel"
          color="grey-7"
          v-close-popup
          :disable="isRunning"
          v-if="!completed"
        />
        <q-btn
          v-if="!isRunning && !completed"
          unelevated
          label="Start Screening"
          color="primary"
          icon="play_arrow"
          @click="startBatchScreening"
        />
        <q-btn
          v-if="completed"
          unelevated
          label="Done & Refresh Grid"
          color="primary"
          v-close-popup
          @click="$emit('batch-finished')"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue'
import api from '../services/api'
import { useQuasar } from 'quasar'

const $q = useQuasar()

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  projectId: {
    type: String,
    required: true
  },
  undecidedCount: {
    type: Number,
    default: 10
  }
})

const emit = defineEmits(['update:modelValue', 'batch-finished'])

const batchLimit = ref(5)
const unscreenedOnly = ref(true)
const selectedProvider = ref('groq')
const customApiKey = ref('')

const isRunning = ref(false)
const completed = ref(false)
const batchResults = ref([])

const providerOptions = [
  { label: 'Groq (Llama 3.3)', value: 'groq' },
  { label: 'OpenAI (GPT-4o-mini)', value: 'openai' },
  { label: 'Google Gemini (1.5 Flash)', value: 'gemini' },
  { label: 'Anthropic Claude (3.5 Haiku)', value: 'claude' }
]

async function startBatchScreening() {
  isRunning.value = true
  completed.value = false
  batchResults.value = []

  try {
    const payload = {
      limit: batchLimit.value,
      undecidedOnly: true,
      unscreenedOnly: unscreenedOnly.value,
      provider: selectedProvider.value,
      apiKey: customApiKey.value
    }

    const res = await api.batchAISuggest(props.projectId, payload)
    batchResults.value = res.data.results || []
    completed.value = true

    $q.notify({
      type: 'positive',
      message: `Batch screening finished! Screened ${res.data.processedCount} studies.`
    })
  } catch (err) {
    $q.notify({
      type: 'negative',
      message: err.response?.data?.error || err.message || 'Batch screening failed'
    })
  } finally {
    isRunning.value = false
  }
}
</script>
