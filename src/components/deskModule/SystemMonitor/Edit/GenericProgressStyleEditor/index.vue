<script setup lang="ts">
import { computed } from 'vue'
import { NColorPicker, NForm, NFormItem, NInputNumber } from 'naive-ui'
import type { GenericProgressStyleExtendParam } from '../../typings'
import GenericMonitorCard from '../../components/GenericMonitorCard/index.vue'
import GenericProgress from '../../components/GenericProgress/index.vue'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { defautSwatchesBackground } from '@/utils/defaultData'

interface Emit {
  (e: 'update:genericProgressStyleExtendParam', visible: GenericProgressStyleExtendParam): void
}

const props = withDefaults(defineProps<{
  genericProgressStyleExtendParam: GenericProgressStyleExtendParam
  /** 阈值单位：percent = 百分比，speed = KB/s（网络卡） */
  thresholdUnit?: 'percent' | 'speed'
}>(), {
  thresholdUnit: 'percent',
})
const emit = defineEmits<Emit>()

// 兼容旧配置（无阈值字段）
const threshold = computed(() => props.genericProgressStyleExtendParam.warningThreshold ?? 0)
const warningColor = computed(() => props.genericProgressStyleExtendParam.warningColor || '#ef4444')
// 预览环固定 50%：百分比模式直接用阈值比较；速度模式只要启用就演示警告色（40 < 50）
const previewThreshold = computed(() => {
  if (threshold.value <= 0)
    return 51
  return props.thresholdUnit === 'speed' ? 40 : threshold.value
})

function updateThreshold(value: number | null) {
  data.value = { ...data.value, warningThreshold: value ?? 0 }
}
function updateWarningColor(value: string) {
  data.value = { ...data.value, warningColor: value }
}

const data = computed({
  get: () => props.genericProgressStyleExtendParam,
  set: (visible) => {
    emit('update:genericProgressStyleExtendParam', visible)
  },
})
</script>

<template>
  <div>
    <!-- <div>{{ genericProgressStyleExtendParam }}</div>
    <div>{{ data }}</div> -->
    <div class="flex mb-5 justify-center transparent-grid p-2 rounded-xl border">
      <div class="w-[200px]">
        <GenericMonitorCard
          icon-text-icon-hide-title
          :card-type-style="PanelPanelConfigStyleEnum.info"
          icon="solar-cpu-bold"
          :background-color="data.backgroundColor"
          :text-color="data.color"
        >
          <template #info>
            <GenericProgress
              :progress-color="data.progressColor"
              :progress-rail-color="data.progressRailColor"
              :percentage="50"
              :progress-height="5"
              info-card-left-text="DEMO"
              info-card-right-text="TEXT"
              :text-color="data.color"
              :card-type-style="PanelPanelConfigStyleEnum.info"
              :warning-threshold="previewThreshold"
              :warning-color="warningColor"
            />
          </template>
        </GenericMonitorCard>
      </div>

      <div class="w-[80px] ml-2">
        <GenericMonitorCard
          icon-text-icon-hide-title
          :card-type-style="PanelPanelConfigStyleEnum.icon"
          icon="solar-cpu-bold"
          :background-color="data.backgroundColor"
          :icon-text-color="data.color"
        >
          <template #small>
            <GenericProgress
              :progress-color="data.progressColor"
              :progress-rail-color="data.progressRailColor"
              :percentage="50"
              :text-color="data.color"
              :card-type-style="PanelPanelConfigStyleEnum.icon"
            />
          </template>
        </GenericMonitorCard>
      </div>
    </div>

    <NForm ref="formRef" v-model="data">
      <NFormItem :label="$t('deskModule.systemMonitor.progressColor')">
        <NColorPicker v-model:value="data.progressColor" :swatches="defautSwatchesBackground" :modes="['hex']" size="small" />
      </NFormItem>
      <NFormItem :label="$t('deskModule.systemMonitor.progressRailColor')">
        <NColorPicker v-model:value="data.progressRailColor" :swatches="defautSwatchesBackground" :modes="['hex']" size="small" />
      </NFormItem>
      <NFormItem :label="$t('common.textColor')">
        <NColorPicker v-model:value="data.color" :swatches="defautSwatchesBackground" :modes="['hex']" size="small" />
      </NFormItem>
      <NFormItem :label="$t('common.backgroundColor')">
        <NColorPicker v-model:value="data.backgroundColor" :swatches="defautSwatchesBackground" :modes="['hex']" size="small" />
      </NFormItem>
      <NFormItem :label="$t('deskModule.systemMonitor.warningThreshold')">
        <NInputNumber
          :value="threshold" :min="0" :max="thresholdUnit === 'percent' ? 100 : 100000"
          :step="thresholdUnit === 'percent' ? 1 : 100" size="small" style="width: 100%;"
          :placeholder="$t('deskModule.systemMonitor.thresholdPlaceholder')"
          @update:value="updateThreshold"
        />
        <template #feedback>
          <span class="text-xs opacity-60">
            {{ thresholdUnit === 'percent' ? $t('deskModule.systemMonitor.thresholdTipPercent') : $t('deskModule.systemMonitor.thresholdTipSpeed') }}
          </span>
        </template>
      </NFormItem>
      <NFormItem :label="$t('deskModule.systemMonitor.warningColor')">
        <NColorPicker :value="warningColor" :swatches="defautSwatchesBackground" :modes="['hex']" size="small" @update:value="updateWarningColor" />
      </NFormItem>
    </NForm>
  </div>
</template>
