<template>
  <q-card flat bordered class="ai-suggestion-card bg-surface">
    <!-- Header -->
    <div class="card-header q-pa-md row items-center justify-between">
      <div class="row items-center q-gutter-sm">
        <q-icon name="psychology" color="primary" size="24px" />
        <span class="text-subtitle1 text-weight-bold">AI Screening Assistant</span>
      </div>
      <q-btn
        unelevated
        color="primary"
        icon="auto_awesome"
        :label="study?.aiSuggestion ? 'Re-Screen with AI' : 'Run AI Screening'"
        :loading="loading"
        :disable="loading || !study"
        @click="$emit('request-ai-suggest')"
      >
        <q-tooltip>Request structured AI screening suggestion (Hotkey: A)</q-tooltip>
      </q-btn>
    </div>

    <q-separator />

    <!-- Card Body -->
    <div class="q-pa-md">
      <!-- Loading State -->
      <div v-if="loading" class="column items-center justify-center q-py-lg text-center">
        <q-spinner-orbit color="primary" size="48px" />
        <div class="text-subtitle2 text-weight-medium q-mt-md text-primary">
          Analyzing Abstract with Protocol Criteria...
        </div>
        <div class="text-caption text-grey-7">
          Evaluating population, intervention, outcomes, and exclusion rules
        </div>
      </div>

      <!-- No AI Screening Result Yet -->
      <div v-else-if="!hasAISuggestion" class="column items-center justify-center q-py-md text-center text-grey-6">
        <q-icon name="insights" size="42px" class="q-mb-xs opacity-60" />
        <div class="text-body2 text-weight-medium">No AI suggestion generated for this study yet</div>
        <div class="text-caption text-grey-7 q-mt-xs">
          Click <strong>Run AI Screening</strong> or press <strong>[A]</strong> to generate real-time LLM recommendations
        </div>
      </div>

      <!-- AI Result Content -->
      <div v-else class="column q-gutter-y-md">
        <!-- Top Status Banner -->
        <div
          class="row items-center justify-between q-pa-md rounded-borders result-banner"
          :class="isInclude ? 'bg-positive-light text-positive-dark' : 'bg-negative-light text-negative-dark'"
        >
          <div class="row items-center q-gutter-sm">
            <q-icon
              :name="isInclude ? 'check_circle' : 'cancel'"
              size="28px"
              :color="isInclude ? 'positive' : 'negative'"
            />
            <div>
              <div class="text-subtitle1 text-weight-bolder text-uppercase">
                AI Suggestion: {{ study.aiSuggestion }}
              </div>
              <div class="text-caption">
                {{ isInclude ? 'Matches protocol inclusion criteria' : 'Violates inclusion or triggers exclusion' }}
              </div>
            </div>
          </div>

          <!-- Confidence Meter -->
          <div class="column items-end">
            <div class="text-caption text-weight-medium text-grey-7">Confidence Score</div>
            <div class="row items-center q-gutter-xs">
              <q-circular-progress
                :value="confidencePercent"
                size="42px"
                :thickness="0.22"
                :color="confidenceColor"
                track-color="grey-3"
                class="q-mr-xs"
              >
                <span class="text-caption text-weight-bold">{{ confidencePercent }}%</span>
              </q-circular-progress>
            </div>
          </div>
        </div>

        <!-- AI Reasoning -->
        <div>
          <div class="text-caption text-weight-bold text-uppercase text-grey-7 q-mb-xs">
            Screening Rationale & Analysis
          </div>
          <div class="reasoning-box q-pa-sm rounded-borders text-body2 bg-grey-1 text-grey-9">
            <q-icon name="format_quote" color="primary" class="q-mr-xs" />
            <span>{{ study.aiReason || study.aiJsonResponse?.reasoning }}</span>
          </div>
        </div>

        <!-- Matched Protocol Criteria -->
        <div v-if="matchedCriteria.length > 0">
          <div class="text-caption text-weight-bold text-uppercase text-grey-7 q-mb-xs">
            Matched Protocol Criteria
          </div>
          <div class="row q-gutter-xs wrap">
            <q-chip
              v-for="(match, idx) in matchedCriteria"
              :key="idx"
              dense
              icon="verified"
              :color="isInclude ? 'green-1' : 'orange-1'"
              :text-color="isInclude ? 'green-9' : 'orange-9'"
              class="text-caption text-weight-medium"
            >
              {{ match }}
            </q-chip>
          </div>
        </div>

        <!-- Collapsible Raw JSON Response for Auditability -->
        <q-expansion-item
          dense
          dense-toggle
          icon="data_object"
          label="View Structured LLM Response (JSON)"
          header-class="text-caption text-grey-8"
        >
          <div class="q-pa-xs">
            <pre class="json-viewer bg-dark text-green-3 q-pa-sm rounded-borders text-caption">{{ formattedJSON }}</pre>
          </div>
        </q-expansion-item>
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  study: {
    type: Object,
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['request-ai-suggest'])

const hasAISuggestion = computed(() => {
  return props.study && (props.study.aiSuggestion || props.study.aiJsonResponse)
})

const isInclude = computed(() => {
  if (!props.study?.aiSuggestion) return false
  return props.study.aiSuggestion.toLowerCase() === 'include'
})

const confidencePercent = computed(() => {
  if (!props.study) return 0
  const conf = props.study.aiConfidence ?? props.study.aiJsonResponse?.confidence ?? 0.85
  return Math.round(conf * 100)
})

const confidenceColor = computed(() => {
  if (confidencePercent.value >= 80) return 'positive'
  if (confidencePercent.value >= 60) return 'warning'
  return 'negative'
})

const matchedCriteria = computed(() => {
  if (!props.study) return []
  return props.study.aiMatches || props.study.aiJsonResponse?.matches || []
})

const formattedJSON = computed(() => {
  if (!props.study?.aiJsonResponse) return '{}'
  return JSON.stringify(props.study.aiJsonResponse, null, 2)
})
</script>

<style scoped>
.ai-suggestion-card {
  border-radius: 8px;
  background: white;
}
.bg-positive-light {
  background-color: #ecfdf5;
  border: 1px solid #a7f3d0;
}
.text-positive-dark {
  color: #065f46;
}
.bg-negative-light {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
}
.text-negative-dark {
  color: #991b1b;
}
.reasoning-box {
  line-height: 1.5;
  border-left: 3px solid #0284c7;
}
.json-viewer {
  font-family: 'JetBrains Mono', monospace;
  max-height: 180px;
  overflow-y: auto;
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
