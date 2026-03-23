<template>
  <div class="overflow-hidden rounded-2xl border border-amber-200 bg-white/90 shadow-sm dark:border-amber-900/40 dark:bg-dark-800/90">
    <div class="border-b border-amber-100 px-4 py-4 dark:border-amber-900/30">
      <div class="text-sm font-semibold text-gray-900 dark:text-white">Nano Banana 定价</div>
      <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">先在模型限制里添加请求模型，再为每个模型设置单张价格。</div>
    </div>

    <div class="p-4">
      <div v-if="models.length === 0" class="rounded-xl border border-dashed border-amber-200 bg-amber-50/60 px-4 py-8 text-center text-sm text-amber-700 dark:border-amber-900/30 dark:bg-amber-950/20 dark:text-amber-300">
        暂无模型。请先添加 Nano Banana 模型。
      </div>

      <div v-else class="overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
        <table class="min-w-full divide-y divide-gray-200 text-xs dark:divide-dark-600">
          <thead class="bg-gray-50 dark:bg-dark-900/40">
            <tr>
              <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">模型</th>
              <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">单张价格</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="model in models" :key="model">
              <td class="px-3 py-2 text-gray-900 dark:text-white">{{ model }}</td>
              <td class="px-3 py-2">
                <div class="flex items-center justify-end gap-2">
                  <span class="text-gray-500 dark:text-gray-400">$</span>
                  <input
                    :value="modelValue[model]?.image_price_per_image ?? ''"
                    type="number"
                    min="0"
                    step="0.0001"
                    class="input h-8 w-28 text-right text-xs"
                    placeholder="0.0000"
                    @input="updatePrice(model, ($event.target as HTMLInputElement).value)"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  models: string[]
  modelValue: Record<string, { image_price_per_image?: number }>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, { image_price_per_image?: number }>): void
}>()

const roundPrice = (value: number) => Math.round(value * 10000) / 10000

const updatePrice = (model: string, value: string) => {
  const parsed = Number.parseFloat(value)
  const next = { ...props.modelValue }
  if (!Number.isFinite(parsed) || parsed <= 0) {
    delete next[model]
    emit('update:modelValue', next)
    return
  }
  next[model] = { image_price_per_image: roundPrice(parsed) }
  emit('update:modelValue', next)
}
</script>
