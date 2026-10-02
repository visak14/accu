<template>
  <q-expansion-item
    dense
    icon="menu_book"
    label="Project Review Protocol & Screening Criteria"
    header-class="bg-blue-grey-1 text-weight-bold text-blue-grey-9 rounded-borders"
    class="protocol-viewer-card q-mb-md"
    default-opened
  >
    <q-card flat bordered class="q-pa-md bg-grey-1">
      <!-- Key Criteria Chips -->
      <div v-if="criteria.length > 0" class="q-mb-sm">
        <div class="text-caption text-weight-bold text-uppercase text-grey-8 q-mb-xs">
          Key Required Criteria
        </div>
        <div class="row q-gutter-xs wrap">
          <q-chip
            v-for="(crit, idx) in criteria"
            :key="idx"
            dense
            color="primary"
            text-color="white"
            icon="checklist"
            class="text-caption"
          >
            {{ crit }}
          </q-chip>
        </div>
      </div>

      <div class="row q-col-gutter-sm q-mt-xs">
        <!-- Inclusion Criteria -->
        <div class="col-12 col-md-6">
          <div class="protocol-box inclusion-box q-pa-sm rounded-borders">
            <div class="row items-center q-gutter-xs text-positive text-weight-bold text-caption q-mb-xs">
              <q-icon name="add_circle" size="16px" />
              <span>INCLUSION CRITERIA</span>
            </div>
            <div class="text-caption text-grey-9 protocol-text">
              {{ inclusionCriteria || 'Primary empirical research matching project scope.' }}
            </div>
          </div>
        </div>

        <!-- Exclusion Criteria -->
        <div class="col-12 col-md-6">
          <div class="protocol-box exclusion-box q-pa-sm rounded-borders">
            <div class="row items-center q-gutter-xs text-negative text-weight-bold text-caption q-mb-xs">
              <q-icon name="remove_circle" size="16px" />
              <span>EXCLUSION CRITERIA</span>
            </div>
            <div class="text-caption text-grey-9 protocol-text">
              {{ exclusionCriteria || 'Non-empirical papers, review articles, animal models, non-target populations.' }}
            </div>
          </div>
        </div>
      </div>
    </q-card>
  </q-expansion-item>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  protocol: {
    type: Object,
    default: () => ({})
  }
})

const criteria = computed(() => props.protocol?.criteria || [])
const inclusionCriteria = computed(() => props.protocol?.inclusion_criteria || '')
const exclusionCriteria = computed(() => props.protocol?.exclusion_criteria || '')
</script>

<style scoped>
.protocol-viewer-card {
  border-radius: 8px;
  overflow: hidden;
}
.protocol-box {
  background: white;
  border: 1px solid #e2e8f0;
  min-height: 80px;
}
.inclusion-box {
  border-left: 3px solid #10b981;
}
.exclusion-box {
  border-left: 3px solid #ef4444;
}
.protocol-text {
  line-height: 1.4;
  white-space: pre-line;
}
</style>
