<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="rounded-3xl border border-gray-200 bg-gradient-to-br from-white via-slate-50 to-emerald-50 p-6 shadow-sm dark:border-dark-600 dark:from-dark-800 dark:via-dark-800 dark:to-emerald-950/20">
        <div class="flex flex-col gap-6 xl:flex-row xl:items-start xl:justify-between">
          <div class="max-w-2xl">
            <div class="inline-flex items-center gap-2 rounded-full bg-white/80 px-3 py-1 text-xs font-medium text-emerald-700 shadow-sm dark:bg-dark-700/80 dark:text-emerald-300">
              <Icon name="search" size="sm" :stroke-width="2" />
              {{ t('modelTest.badge') }}
            </div>
            <h1 class="mt-4 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('modelTest.title') }}
            </h1>
            <p class="mt-3 max-w-xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('modelTest.description') }}
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-3 xl:w-[360px] xl:grid-cols-1">
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.keyCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ apiKeys.length }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.modelCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ models.length }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.pricedCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ pricedModelsCount }}</div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
        <aside class="space-y-6">
          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.keyPanel.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.keyPanel.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loadingBootstrap" @click="loadBootstrap">
                {{ t('common.refresh') }}
              </button>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('modelTest.keyPanel.existingKeys') }}</label>
                <Select
                  v-model="selectedApiKeyId"
                  :options="apiKeyOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('modelTest.keyPanel.existingKeysPlaceholder')"
                  @change="applySelectedApiKey"
                />
              </div>

              <div>
                <label class="input-label">{{ t('modelTest.keyPanel.directInput') }}</label>
                <div class="flex gap-2">
                  <input v-model="apiKeyInput" type="password" class="input flex-1 font-mono" :placeholder="t('modelTest.keyPanel.directInputPlaceholder')" />
                  <button type="button" class="btn btn-secondary" :disabled="!apiKeyInput.trim()" @click="loadModels">
                    {{ t('modelTest.keyPanel.verify') }}
                  </button>
                </div>
              </div>

              <div class="rounded-xl border border-emerald-200 bg-emerald-50/70 p-4 dark:border-emerald-900/40 dark:bg-emerald-950/20">
                <div class="text-sm font-medium text-emerald-900 dark:text-emerald-200">{{ t('modelTest.keyPanel.generateTitle') }}</div>
                <div class="mt-1 text-xs text-emerald-700 dark:text-emerald-300">{{ t('modelTest.keyPanel.generateHint') }}</div>

                <div class="mt-4 space-y-3">
                  <div>
                    <label class="input-label">{{ t('modelTest.keyPanel.groupLabel') }}</label>
                    <Select
                      v-model="selectedGroupId"
                      :options="groupOptions"
                      value-key="value"
                      label-key="label"
                      :placeholder="t('modelTest.keyPanel.groupPlaceholder')"
                    />
                  </div>
                  <button type="button" class="btn btn-primary w-full" :disabled="generatingKey || selectedGroupId == null" @click="generateApiKey">
                    {{ generatingKey ? t('common.loading') + '...' : t('modelTest.keyPanel.generateButton') }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.models.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.models.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loadingModels || !apiKeyInput.trim()" @click="loadModels">
                {{ loadingModels ? t('common.loading') + '...' : t('modelTest.models.fetch') }}
              </button>
            </div>

            <div class="mt-4 max-h-[360px] overflow-y-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <div v-if="models.length === 0" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('modelTest.models.empty') }}
              </div>
              <button
                v-for="model in models"
                :key="model.id"
                type="button"
                class="flex w-full items-center justify-between border-b border-gray-100 px-4 py-3 text-left transition hover:bg-gray-50 last:border-b-0 dark:border-dark-700 dark:hover:bg-dark-700/50"
                @click="selectModel(model.id)"
              >
                <div class="min-w-0">
                  <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ model.display_name || model.id }}</div>
                  <div class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ model.id }}</div>
                </div>
                <span class="rounded-full bg-gray-100 px-2 py-1 text-[11px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">{{ t('common.select') }}</span>
              </button>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">生成历史</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">最近的提示词记录</p>
              </div>
              <button type="button" class="btn btn-secondary text-xs" @click="clearHistory">
                清空
              </button>
            </div>

            <div class="mt-4 max-h-[240px] space-y-2 overflow-y-auto">
              <div v-if="promptHistory.length === 0" class="rounded-xl border border-dashed border-gray-300 px-4 py-6 text-center text-xs text-gray-500 dark:border-dark-600 dark:text-gray-400">
                暂无历史记录
              </div>
              <button
                v-for="(item, index) in promptHistory"
                :key="`history-${index}`"
                type="button"
                @click="applyHistoryItem(item)"
                class="w-full rounded-lg border border-gray-200 bg-gray-50 p-3 text-left text-xs transition hover:border-primary-500 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-900/30 dark:hover:border-primary-500 dark:hover:bg-primary-950/30"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="min-w-0 flex-1">
                    <div class="mb-1 flex items-center gap-2">
                      <span class="rounded bg-emerald-100 px-2 py-0.5 text-[10px] font-medium text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-400">
                        {{ item.type === 'image' ? '图片' : item.type === 'video' ? '视频' : '编辑' }}
                      </span>
                      <span class="text-[10px] text-gray-400">{{ formatHistoryTime(item.timestamp) }}</span>
                    </div>
                    <div class="line-clamp-2 text-gray-700 dark:text-gray-300">{{ item.prompt }}</div>
                    <div v-if="item.model" class="mt-1 text-[10px] text-gray-400">{{ item.model }}</div>
                  </div>
                  <button
                    type="button"
                    @click.stop="removeHistoryItem(index)"
                    class="rounded p-1 text-gray-400 hover:bg-red-100 hover:text-red-600 dark:hover:bg-red-950/50"
                  >
                    <Icon name="x" size="xs" />
                  </button>
                </div>
              </button>
            </div>
          </section>
        </aside>

        <section class="space-y-6">
          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.pricing.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.description') }}</p>
              </div>
              <div class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                {{ effectiveRateLabel }}
              </div>
            </div>

            <div class="mt-4 overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <table class="min-w-full divide-y divide-gray-200 text-xs dark:divide-dark-600">
                <thead class="bg-gray-50 dark:bg-dark-900/40">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.name') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.standardInput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.standardOutput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.actualInput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.actualOutput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">图片/视频单次价格</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-if="models.length === 0">
                    <td colspan="6" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.empty') }}</td>
                  </tr>
                  <tr v-for="model in pricedModels" :key="model.id">
                    <td class="px-3 py-2 text-gray-900 dark:text-white">{{ model.id }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.standardInputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.standardOutputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.actualInputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.actualOutputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatImageOrVideoPrice(model.id, model.imagePricePerImage, model.pricingAvailable && model.imagePricePerImage > 0) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">图片处理模型</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">无参考图时生图，上传参考图后可做参考图生成或改图；Nano Banana 会直接走 `/v1/draw/nano-banana`。</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[560px]">
                <Select v-model="imageModel" class="flex-1" :options="imageModelOptions" value-key="value" label-key="label" :placeholder="t('modelTest.image.selectModel')" />
                <Select v-if="!isNanoBananaImageModel" v-model="imageSize" class="w-[140px]" :options="imageSizeOptions" value-key="value" label-key="label" />
                <Select v-else v-model="nanoBananaAspectRatio" class="w-[120px]" :options="nanoBananaAspectRatioSelectOptionsForGenerate" value-key="value" label-key="label" />
                <Select v-if="isNanoBananaImageModel" v-model="nanoBananaImageSize" class="w-[110px]" :options="nanoBananaImageSizeOptionsForGenerate" value-key="value" label-key="label" />
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <textarea v-model="imagePrompt" rows="3" class="input" :placeholder="t('modelTest.image.placeholder')" />
              <div v-if="isNanoBananaImageModel" class="rounded-xl border border-amber-200 bg-amber-50/70 px-3 py-2 text-xs text-amber-800 dark:border-amber-900/30 dark:bg-amber-950/20 dark:text-amber-300">
                Nano Banana 使用专属绘图接口，默认优先选择 2K；如果当前模型不支持 2K，会自动切到该模型允许的分辨率。
              </div>

              <div v-if="generatingImage && isNanoBananaImageModel" class="space-y-2 rounded-xl border border-emerald-200 bg-emerald-50/70 px-3 py-3 dark:border-emerald-900/30 dark:bg-emerald-950/20">
                <div class="flex items-center justify-between gap-3 text-xs">
                  <span class="font-medium text-emerald-800 dark:text-emerald-300">{{ imageProcessingStatus || '处理中...' }}</span>
                  <span class="text-emerald-700 dark:text-emerald-400">{{ imageProcessingProgress }}%</span>
                </div>
                <div class="h-2 overflow-hidden rounded-full bg-emerald-100 dark:bg-emerald-900/40">
                  <div class="h-full rounded-full bg-emerald-500 transition-all duration-300" :style="{ width: `${imageProcessingProgress}%` }" />
                </div>
              </div>

              <div>
                <label class="input-label">参考图 / 待处理图（可选）</label>
                <div class="space-y-3">
                  <input
                    ref="imageReferenceFileInput"
                    type="file"
                    accept="image/*"
                    multiple
                    class="hidden"
                    @change="handleImageReferenceFileChange"
                  />
                  <div
                    @drop.prevent="handleImageReferenceDrop"
                    @dragover.prevent="imageReferenceDragover = true"
                    @dragleave.prevent="imageReferenceDragover = false"
                    :class="[
                      'overflow-hidden rounded-xl border-2 border-dashed transition',
                      imageReferenceDragover
                        ? 'border-amber-400 bg-amber-50 dark:border-amber-500 dark:bg-amber-950/20'
                        : 'border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-900/30'
                    ]"
                  >
                    <button
                      type="button"
                      class="flex w-full items-center justify-center px-4 py-7 text-sm text-gray-600 dark:text-gray-400"
                      @click="imageReferenceFileInput?.click()"
                    >
                      <Icon name="upload" size="sm" class="mr-2" />
                      {{ imageReferenceItems.length === 0 ? '点击或拖拽上传图片（Nano Banana 支持多张）' : '继续添加图片' }}
                    </button>
                  </div>

                  <div v-if="imageReferenceItems.length > 0" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                    <div
                      v-for="(item, index) in imageReferenceItems"
                      :key="item.key"
                      class="group relative overflow-hidden rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-900/30"
                    >
                      <img :src="item.preview" :alt="`reference-${index}`" class="aspect-square w-full object-cover" />
                      <div class="absolute left-2 top-2 rounded-full bg-black/60 px-2 py-0.5 text-[11px] font-medium text-white">
                        #{{ index + 1 }}
                      </div>
                      <button
                        type="button"
                        @click="removeImageReference(index)"
                        class="absolute right-2 top-2 rounded-lg bg-red-500 px-2.5 py-1 text-xs font-medium text-white shadow-lg transition hover:bg-red-600"
                      >
                        移除
                      </button>
                    </div>
                  </div>

                  <div v-if="imageReferenceItems.length > 0" class="flex items-center justify-between gap-3 rounded-xl border border-amber-200 bg-amber-50/70 px-3 py-2 text-xs text-amber-800 dark:border-amber-900/30 dark:bg-amber-950/20 dark:text-amber-300">
                    <span v-if="isNanoBananaImageModel">已添加 {{ imageReferenceItems.length }} 张参考图，将一起传给 Nano Banana。</span>
                    <span v-else>当前非 Nano 模型仅支持普通生图；如需参考图处理，请切换到 Nano Banana 系列模型。</span>
                    <button type="button" class="font-medium underline underline-offset-2" @click="clearImageReferencePreview">
                      清空全部
                    </button>
                  </div>
                </div>
              </div>
               
              <!-- 风格标签 -->
              <div>
                <label class="input-label">风格标签（点击追加到提示词）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="style in imageStyles"
                    :key="style"
                    type="button"
                    @click="appendToImagePrompt(style)"
                    class="rounded-lg border border-gray-300 bg-white px-3 py-1 text-xs font-medium text-gray-700 transition hover:border-primary-500 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500 dark:hover:bg-primary-950/50 dark:hover:text-primary-400"
                  >
                    {{ style }}
                  </button>
                </div>
              </div>

              <!-- 示例提示词 -->
              <div>
                <label class="input-label">快捷模板（点击替换提示词）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="example in imageExamples"
                    :key="example"
                    type="button"
                    @click="imagePrompt = example"
                    class="rounded-lg border border-gray-300 bg-white px-3 py-1 text-xs font-medium text-gray-700 transition hover:border-primary-500 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500 dark:hover:bg-primary-950/50 dark:hover:text-primary-400"
                  >
                    {{ example.split('，')[0] }}...
                  </button>
                </div>
              </div>
              
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  <span v-if="estimatedImageCost > 0 || estimatedImageDurationLabel">{{ [estimatedImageCostLabel, estimatedImageDurationLabel].filter(Boolean).join(' · ') }}</span>
                  <span v-else>{{ t('modelTest.image.hint') }}</span>
                  <span v-if="imageLastDurationLabel" class="block mt-1 text-emerald-600 dark:text-emerald-400">{{ imageLastDurationLabel }}</span>
                </div>
                <button type="button" class="btn btn-primary" :disabled="generatingImage || !apiKeyInput.trim() || !imageModel || !imagePrompt.trim()" @click="generateImage">
                  {{ generatingImage ? t('common.loading') + '...' : '开始处理' }}
                </button>
              </div>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <div v-if="generatedImages.length === 0" class="col-span-full rounded-xl border border-dashed border-gray-300 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
                暂无处理结果
              </div>
              <div v-for="(image, index) in generatedImages" :key="`${image}-${index}`" class="group relative overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                <img :src="image" :alt="`generated-${index}`" class="aspect-square w-full object-cover" />
                <div class="absolute inset-x-0 bottom-0 flex gap-1 bg-gradient-to-t from-black/80 to-transparent p-3 opacity-0 transition group-hover:opacity-100">
                  <button type="button" @click="downloadImage(image, `generated-${index}.png`)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    下载
                  </button>
                  <button type="button" @click="copyImageUrl(image)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    复制
                  </button>
                  <button type="button" @click="removeGeneratedImage(index)" class="rounded-lg bg-red-500/90 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-red-500">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-show="false" class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">图片编辑</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">上传图片并使用 AI 进行编辑</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[560px]">
                <Select v-model="imageEditModel" class="flex-1" :options="imageEditModelOptions" value-key="value" label-key="label" placeholder="选择编辑模型" />
                <Select v-if="!isNanoBananaEditModel" v-model="imageEditSize" class="w-[140px]" :options="imageSizeOptions" value-key="value" label-key="label" />
                <Select v-else v-model="nanoBananaEditAspectRatio" class="w-[120px]" :options="nanoBananaAspectRatioSelectOptionsForEdit" value-key="value" label-key="label" />
                <Select v-if="isNanoBananaEditModel" v-model="nanoBananaEditImageSize" class="w-[110px]" :options="nanoBananaImageSizeOptionsForEdit" value-key="value" label-key="label" />
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <!-- 图片上传区域（支持拖拽） -->
              <div>
                <label class="input-label">上传图片</label>
                <div
                  v-if="!imageEditPreview"
                  @drop.prevent="handleImageEditDrop"
                  @dragover.prevent="imageEditDragover = true"
                  @dragleave.prevent="imageEditDragover = false"
                  :class="[
                    'relative overflow-hidden rounded-xl border-2 border-dashed transition',
                    imageEditDragover
                      ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-950/30'
                      : 'border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-900/30'
                  ]"
                >
                  <input
                    ref="imageEditFileInput"
                    type="file"
                    accept="image/*"
                    class="hidden"
                    @change="handleImageEditFileChange"
                  />
                  <button
                    type="button"
                    @click="imageEditFileInput?.click()"
                    class="flex w-full cursor-pointer items-center justify-center px-4 py-8 text-sm text-gray-600 dark:text-gray-400"
                  >
                    <Icon name="upload" size="sm" class="mr-2" />
                    点击或拖拽上传图片（支持 PNG, JPG, WEBP，最大 10MB）
                  </button>
                </div>
                <div v-else class="relative overflow-hidden rounded-xl border border-gray-200 dark:border-dark-600">
                  <img :src="imageEditPreview" alt="Preview" class="w-full" />
                  <button
                    type="button"
                    @click="clearImageEditPreview"
                    class="absolute right-2 top-2 rounded-lg bg-red-500 px-3 py-1.5 text-xs font-medium text-white shadow-lg transition hover:bg-red-600"
                  >
                    重新选择
                  </button>
                </div>
              </div>

              <!-- 常用编辑任务快捷按钮 -->
              <div>
                <label class="input-label">编辑任务</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="task in imageEditTasks"
                    :key="task"
                    type="button"
                    @click="imageEditPrompt = task"
                    :class="[
                      'rounded-lg border px-3 py-2 text-xs font-medium transition',
                      imageEditPrompt === task
                        ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-950/50 dark:text-primary-400'
                        : 'border-gray-300 bg-white text-gray-700 hover:border-primary-400 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500'
                    ]"
                  >
                    {{ task }}
                  </button>
                </div>
              </div>

              <textarea v-model="imageEditPrompt" rows="3" class="input" placeholder="描述你想对图片做什么修改，或使用上方快捷任务..." />
              <div v-if="isNanoBananaEditModel" class="rounded-xl border border-amber-200 bg-amber-50/70 px-3 py-2 text-xs text-amber-800 dark:border-amber-900/30 dark:bg-amber-950/20 dark:text-amber-300">
                Nano Banana 改图会自动把上传图片转换为参考图并走 `/v1/draw/nano-banana`。
              </div>
              
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  <span v-if="estimatedEditCost > 0">{{ estimatedEditCostLabel }}</span>
                  <span v-else>等待输入...</span>
                </div>
                <button type="button" class="btn btn-primary" :disabled="editingImage || !apiKeyInput.trim() || !imageEditModel || !imageEditPrompt.trim() || !imageEditFile" @click="editImage">
                  {{ editingImage ? '编辑中...' : '开始编辑' }}
                </button>
              </div>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <div v-if="editedImages.length === 0" class="col-span-full rounded-xl border border-dashed border-gray-300 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
                暂无编辑结果
              </div>
              <div v-for="(image, index) in editedImages" :key="`edited-${image}-${index}`" class="group relative overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                <img :src="image" :alt="`edited-${index}`" class="aspect-square w-full object-cover" />
                <div class="absolute inset-x-0 bottom-0 flex gap-1 bg-gradient-to-t from-black/80 to-transparent p-3 opacity-0 transition group-hover:opacity-100">
                  <button type="button" @click="downloadImage(image, `edited-${index}.png`)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    下载
                  </button>
                  <button type="button" @click="copyImageUrl(image)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    复制
                  </button>
                  <button type="button" @click="removeEditedImage(index)" class="rounded-lg bg-red-500/90 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-red-500">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">视频生成</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">使用 AI 生成视频内容</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[480px]">
                <Select v-model="videoModel" class="flex-1" :options="videoModelOptions" value-key="value" label-key="label" placeholder="选择视频模型" />
                <Select v-model="videoSize" class="w-[140px]" :options="videoSizeOptions" value-key="value" label-key="label" />
                <Select v-model="videoQuality" class="w-[110px]" :options="videoQualityOptions" value-key="value" label-key="label" />
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <!-- 场景模板 -->
              <div>
                <label class="input-label">场景模板（点击替换提示词）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="scene in videoScenes"
                    :key="scene.prompt"
                    type="button"
                    @click="videoPrompt = scene.prompt"
                    class="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 transition hover:border-primary-500 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500 dark:hover:bg-primary-950/50 dark:hover:text-primary-400"
                  >
                    {{ scene.name }}
                  </button>
                </div>
              </div>

              <textarea v-model="videoPrompt" rows="3" class="input" placeholder="描述你想生成的视频内容，或使用上方场景模板..." />

              <!-- 镜头运动选项 -->
              <div>
                <label class="input-label">镜头运动（点击追加到提示词）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="camera in cameraMovements"
                    :key="camera"
                    type="button"
                    @click="appendToVideoPrompt(camera)"
                    class="rounded-lg border border-gray-300 bg-white px-3 py-1 text-xs font-medium text-gray-700 transition hover:border-primary-500 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500 dark:hover:bg-primary-950/50 dark:hover:text-primary-400"
                  >
                    {{ camera }}
                  </button>
                </div>
              </div>
              
              <!-- Reference Image Upload -->
              <div>
                <label class="input-label">参考图（可选）</label>
                <div v-if="!videoReferenceImage" class="relative">
                  <input
                    type="file"
                    accept="image/*"
                    class="hidden"
                    id="video-reference-upload"
                    @change="handleVideoReferenceImageUpload"
                  />
                  <label
                    for="video-reference-upload"
                    class="flex cursor-pointer items-center justify-center rounded-xl border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-6 text-sm text-gray-600 transition hover:border-primary-400 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-900/30 dark:text-gray-400 dark:hover:border-primary-500 dark:hover:bg-primary-950/30"
                  >
                    <Icon name="upload" size="sm" class="mr-2" />
                    点击上传参考图（支持 JPG/PNG，最大 10MB）
                  </label>
                </div>
                <div v-else class="relative overflow-hidden rounded-xl border border-gray-200 dark:border-dark-600">
                  <img :src="videoReferenceImage" alt="Reference" class="w-full" />
                  <button
                    type="button"
                    @click="removeVideoReferenceImage"
                    class="absolute right-2 top-2 rounded-lg bg-red-500 px-3 py-1.5 text-xs font-medium text-white shadow-lg transition hover:bg-red-600"
                  >
                    移除
                  </button>
                </div>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  上传参考图可以让 AI 根据图片内容生成相关视频
                </p>
              </div>

              <div class="flex items-center gap-4">
                <div class="flex-1">
                  <label class="input-label">视频时长（秒）</label>
                  <input v-model.number="videoSeconds" type="number" min="6" max="30" class="input" />
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 pt-6">
                  6-30 秒
                  <span v-if="videoSeconds > 12" class="block text-yellow-600 dark:text-yellow-400 mt-1">
                    ⏱️ {{ videoSeconds }}秒视频需要等待 {{ Math.ceil(videoSeconds / 6 * 0.5) }}-{{ Math.ceil(videoSeconds / 6) }} 分钟
                  </span>
                </div>
                <div class="flex-1">
                  <label class="input-label">快捷时长</label>
                  <div class="flex gap-2">
                    <button v-for="dur in [6, 12, 20, 30]" :key="dur" type="button" @click="videoSeconds = dur" :class="['flex-1 rounded-lg border px-3 py-2 text-sm font-medium transition', videoSeconds === dur ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-950/50 dark:text-primary-400' : 'border-gray-300 bg-white text-gray-700 hover:border-primary-400 hover:bg-primary-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-primary-500']">
                      {{dur}}s
                    </button>
                  </div>
                </div>
              </div>
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  <span v-if="estimatedVideoCost > 0">预计费用：${{ estimatedVideoCost.toFixed(4) }} · </span>
                  {{ videoProgress || '等待生成...' }}
                </div>
                <button type="button" class="btn btn-primary" :disabled="generatingVideo || !apiKeyInput.trim() || !videoModel || !videoPrompt.trim()" @click="generateVideo">
                  {{ generatingVideo ? '生成中...' : '生成视频' }}
                </button>
              </div>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2">
              <div v-if="generatedVideos.length === 0" class="col-span-full rounded-xl border border-dashed border-gray-300 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
                暂无生成结果
              </div>
              <div v-for="(video, index) in generatedVideos" :key="`video-${video}-${index}`" class="group relative overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                <video :src="video" controls class="w-full" />
                <div class="absolute inset-x-0 bottom-0 flex gap-1 bg-gradient-to-t from-black/80 to-transparent p-3 opacity-0 transition group-hover:opacity-100">
                  <button type="button" @click="downloadVideo(video, `video-${index}.mp4`)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    下载
                  </button>
                  <button type="button" @click="copyVideoUrl(video)" class="flex-1 rounded-lg bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-900 transition hover:bg-white">
                    复制
                  </button>
                  <button type="button" @click="removeGeneratedVideo(index)" class="rounded-lg bg-red-500/90 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-red-500">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.chat.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.chat.description') }}</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[360px]">
                <Select
                  v-model="chatModel"
                  class="flex-1"
                  :options="modelOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('modelTest.chat.selectModel')"
                />
                <button type="button" class="btn btn-secondary" @click="clearChat">{{ t('modelTest.chat.clear') }}</button>
              </div>
            </div>

            <div ref="chatScrollRef" class="mt-4 min-h-[360px] max-h-[520px] overflow-y-auto rounded-2xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-900/30">
              <div v-if="chatMessages.length === 0" class="flex h-[320px] items-center justify-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('modelTest.chat.empty') }}
              </div>
              <div v-for="(message, index) in chatMessages" :key="index" class="mb-3 flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
                <div :class="message.role === 'user' ? 'max-w-[85%] rounded-2xl rounded-br-md bg-emerald-600 px-4 py-3 text-sm text-white' : 'max-w-[85%] rounded-2xl rounded-bl-md bg-white px-4 py-3 text-sm text-gray-800 shadow-sm dark:bg-dark-700 dark:text-gray-100'">
                  <div class="whitespace-pre-wrap break-words">{{ message.content }}</div>
                </div>
              </div>
              <div v-if="sending" class="mb-3 flex justify-start">
                <div class="rounded-2xl rounded-bl-md bg-white px-4 py-3 text-sm text-gray-500 shadow-sm dark:bg-dark-700 dark:text-gray-300">
                  {{ t('modelTest.chat.sending') }}
                </div>
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <textarea
                v-model="chatInput"
                rows="4"
                class="input"
                :placeholder="t('modelTest.chat.placeholder')"
                @keydown.enter.exact.prevent="sendMessage"
              />
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.chat.hint') }}</div>
                <button type="button" class="btn btn-primary" :disabled="sending || !apiKeyInput.trim() || !chatModel || !chatInput.trim()" @click="sendMessage">
                  {{ t('modelTest.chat.send') }}
                </button>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { useImageEditAPI } from '@/composables/useImageEditAPI'
import { useVideoAPI } from '@/composables/useVideoAPI'
import type { ApiKey, Group, ModelPricingPreviewItem } from '@/types'
import {
  getNanoBananaAspectRatioOptions,
  getNanoBananaDefaultImageSize,
  getNanoBananaSupportedImageSizes,
  isNanoBananaModel
} from '@/utils/nanoBanana'

type LiveModel = { id: string; display_name?: string }
type ChatMessage = { role: 'user' | 'assistant'; content: string }
type ImageGenerationResponse = { data?: Array<{ url?: string }> }
type NanoBananaCreateResponse = { code?: number; msg?: string; data?: { id?: string } }
type NanoBananaResultResponse = {
  code?: number
  msg?: string
  data?: {
    start_time?: number
    end_time?: number
    id?: string
    status?: string
    progress?: number
    error?: string
    failure_reason?: string
    results?: Array<{ url?: string }>
  }
}
type ImageReferenceItem = { key: string; file: File; preview: string; base64: string }

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const { editImage: editImageAPI } = useImageEditAPI()
const { generateVideo: generateVideoAPI, progress: videoProgress } = useVideoAPI()

const loadingBootstrap = ref(false)
const loadingModels = ref(false)
const generatingKey = ref(false)
const sending = ref(false)
const generatingImage = ref(false)
const editingImage = ref(false)
const generatingVideo = ref(false)
const imageLastDurationMs = ref<number | null>(null)
const imageProcessingProgress = ref(0)
const imageProcessingStatus = ref('')
let imageProcessingTimer: number | null = null

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const userGroupRates = ref<Record<number, number>>({})
const models = ref<LiveModel[]>([])
const pricingMap = ref<Record<string, ModelPricingPreviewItem>>({})
const selectedApiKeyId = ref<number | null>(null)
const selectedGroupId = ref<number | null>(null)
const apiKeyInput = ref('')
const chatModel = ref('')
const chatInput = ref('')
const chatMessages = ref<ChatMessage[]>([])
const chatScrollRef = ref<HTMLElement | null>(null)
const imageModel = ref('')
const imagePrompt = ref('')
const imageExamples = ref([
  '赛博朋克城市夜景，霓虹灯光，下雨，电影级画质',
  '可爱的橘猫，坐在窗台上，温暖阳光，4K高清',
  '抽象艺术，色彩斑斓，流动感，现代艺术风格',
  '神秘森林，晨雾缭绕，光线透过树叶，魔幻氛围',
  '日式庭院，樱花盛开，池塘锦鲤，宁静和风',
  '科幻太空站，星空背景，高科技设备，未来感'
])

const imageStyles = ref([
  '写实照片风格',
  '动漫插画风格',
  '水彩画风格',
  '油画风格',
  '赛博朋克风格',
  '梵高风格',
  '吉卜力风格',
  '极简主义',
  '蒸汽波美学',
  '3D渲染'
])

const imageEditTasks = ref([
  '提高清晰度和分辨率',
  '去除背景',
  '修复图片瑕疵',
  '改为动漫风格',
  '改为油画风格',
  '增强色彩饱和度',
  '添加艺术效果',
  '改变光线和氛围'
])

const videoScenes = ref([
  { name: '🌃 都市夜景', prompt: '霓虹灯闪烁的城市夜景，雨后街道反光，车流穿梭，赛博朋克氛围' },
  { name: '🌲 神秘森林', prompt: '清晨的神秘森林，阳光穿过树叶，晨雾缭绕，光束效果，魔幻氛围' },
  { name: '🌊 海浪冲击', prompt: '海浪拍打礁石，水花四溅，夕阳余晖，慢动作拍摄' },
  { name: '🚀 科幻场景', prompt: '未来科幻城市，飞行器穿梭，全息投影，高科技建筑，蓝紫色调' },
  { name: '🏔️ 雪山延时', prompt: '雪山日出延时摄影，云海翻腾，金色阳光，壮丽景观' },
  { name: '🎆 烟花绽放', prompt: '璀璨烟花在夜空绽放，色彩斑斓，慢镜头拍摄' },
  { name: '☕ 咖啡特写', prompt: '咖啡倒入杯中特写，奶油花纹形成，蒸汽升腾，温暖色调' },
  { name: '🌸 花朵绽放', prompt: '花朵绽放过程延时摄影，微距镜头，柔和光线，春天氛围' }
])

const cameraMovements = ref([
  '推镜头（Dolly In）',
  '拉镜头（Dolly Out）',
  '横摇（Pan）',
  '竖摇（Tilt）',
  '环绕（Orbit）',
  '升降（Crane）',
  '跟随（Follow）',
  '固定镜头（Static）',
  '慢动作（Slow Motion）',
  '延时摄影（Time-lapse）'
])
const imageSize = ref('1024x1024')
const nanoBananaAspectRatio = ref('auto')
const nanoBananaImageSize = ref('2K')
const imageReferenceItems = ref<ImageReferenceItem[]>([])
const imageReferenceFileInput = ref<HTMLInputElement | null>(null)
const imageReferenceDragover = ref(false)
const generatedImages = ref<string[]>([])

// Image edit states
const imageEditModel = ref('')
const imageEditPrompt = ref('')
const imageEditSize = ref('1024x1024')
const nanoBananaEditAspectRatio = ref('auto')
const nanoBananaEditImageSize = ref('2K')
const imageEditFile = ref<File | null>(null)
const imageEditFileInput = ref<HTMLInputElement | null>(null)
const imageEditPreview = ref<string | null>(null)
const imageEditDragover = ref(false)
const editedImages = ref<string[]>([])

// Video generation states
const videoModel = ref('')
const videoPrompt = ref('')
const videoSize = ref('1280x720')
const videoQuality = ref<'standard' | 'high'>('standard')
const videoSeconds = ref(6)
const generatedVideos = ref<string[]>([])
const videoReferenceImage = ref<string | null>(null)
const videoReferenceImageFile = ref<File | null>(null)

// History
type HistoryItem = {
  type: 'image' | 'video' | 'edit'
  prompt: string
  model: string
  timestamp: number
  size?: string
  quality?: string
  seconds?: number
}
const promptHistory = ref<HistoryItem[]>([])
const MAX_HISTORY = 20

// Load history from localStorage
const loadHistory = () => {
  try {
    const saved = localStorage.getItem('model-test-history')
    if (saved) {
      promptHistory.value = JSON.parse(saved)
    }
  } catch (error) {
    console.error('Failed to load history:', error)
  }
}

// Save history to localStorage
const saveHistory = () => {
  try {
    localStorage.setItem('model-test-history', JSON.stringify(promptHistory.value))
  } catch (error) {
    console.error('Failed to save history:', error)
  }
}

// Add to history
const addToHistory = (item: HistoryItem) => {
  promptHistory.value.unshift(item)
  if (promptHistory.value.length > MAX_HISTORY) {
    promptHistory.value = promptHistory.value.slice(0, MAX_HISTORY)
  }
  saveHistory()
}

// Format time
const formatHistoryTime = (timestamp: number) => {
  const now = Date.now()
  const diff = now - timestamp
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return new Date(timestamp).toLocaleDateString('zh-CN')
}

// Apply history item
const applyHistoryItem = (item: HistoryItem) => {
  if (item.type === 'image') {
    imagePrompt.value = item.prompt
    if (item.model) imageModel.value = item.model
    if (item.size) imageSize.value = item.size
  } else if (item.type === 'video') {
    videoPrompt.value = item.prompt
    if (item.model) videoModel.value = item.model
    if (item.size) videoSize.value = item.size
    if (item.quality) videoQuality.value = item.quality as 'standard' | 'high'
    if (item.seconds) videoSeconds.value = item.seconds
  } else if (item.type === 'edit') {
    imageEditPrompt.value = item.prompt
    if (item.model) imageEditModel.value = item.model
    if (item.size) imageEditSize.value = item.size
  }
  appStore.showSuccess('已应用历史记录')
}

// Clear history
const clearHistory = () => {
  promptHistory.value = []
  saveHistory()
  appStore.showSuccess('历史记录已清空')
}

// Remove history item
const removeHistoryItem = (index: number) => {
  promptHistory.value.splice(index, 1)
  saveHistory()
}

const apiKeyOptions = computed(() => apiKeys.value.map((key) => ({ value: key.id, label: `${key.name} · ${maskKey(key.key)}` })))
const groupOptions = computed(() => groups.value.map((group) => ({ value: group.id, label: `${group.name} · ${normalizeMultiplierValue(effectiveRateForGroup(group.id)).toFixed(2)}x` })))
const modelOptions = computed(() => models.value.map((model) => ({ value: model.id, label: model.display_name || model.id })))
const imageModelOptions = computed(() =>
  models.value
    .filter((model) => {
      const pricing = pricingMap.value[model.id]
      if ((pricing?.image_price_per_image || 0) > 0) {
        return true
      }
      return isNanoBananaModel(model.id) || /(imagine|image|img|flux|sdxl|dall-e|recraft|canvas|vision-image)/i.test(model.id)
    })
    .map((model) => ({ value: model.id, label: model.display_name || model.id }))
)
const imageSizeOptions = [
  { value: '1024x1024', label: '1024x1024' },
  { value: '1792x1024', label: '1792x1024' },
  { value: '1024x1792', label: '1024x1792' }
]

const imageEditModelOptions = computed(() =>
  models.value
    .filter((model) => isNanoBananaModel(model.id) || /(imagine.*edit|edit|grok.*edit)/i.test(model.id))
    .map((model) => ({ value: model.id, label: model.display_name || model.id }))
)

const nanoBananaImageSizeOptionsForGenerate = computed(() =>
  getNanoBananaSupportedImageSizes(imageModel.value).map((value) => ({ value, label: value }))
)

const nanoBananaImageSizeOptionsForEdit = computed(() =>
  getNanoBananaSupportedImageSizes(imageEditModel.value).map((value) => ({ value, label: value }))
)

const nanoBananaAspectRatioSelectOptionsForGenerate = computed(() =>
  getNanoBananaAspectRatioOptions(imageModel.value).map((value) => ({ value, label: value }))
)

const nanoBananaAspectRatioSelectOptionsForEdit = computed(() =>
  getNanoBananaAspectRatioOptions(imageEditModel.value).map((value) => ({ value, label: value }))
)

const videoModelOptions = computed(() =>
  models.value
    .filter((model) => /(video|sora|grok.*video)/i.test(model.id))
    .map((model) => ({ value: model.id, label: model.display_name || model.id }))
)

const videoSizeOptions = [
  { value: '1280x720', label: '1280x720 (16:9)' },
  { value: '720x1280', label: '720x1280 (9:16)' },
  { value: '1792x1024', label: '1792x1024 (宽屏)' },
  { value: '1024x1792', label: '1024x1792 (竖屏)' },
  { value: '1024x1024', label: '1024x1024 (方形)' }
]

const videoQualityOptions = [
  { value: 'standard', label: '标准 (480p)' },
  { value: 'high', label: '高清 (720p)' }
]

const effectiveRateForGroup = (groupId?: number | null) => {
  if (!groupId) return 1
  const group = groups.value.find((item) => item.id === groupId) || apiKeys.value.find((item) => item.group_id === groupId)?.group
  const userRate = userGroupRates.value[groupId]
  if (userRate != null && userRate > 0) return userRate
  return group?.rate_multiplier || 1
}

const activeApiKey = computed(() => {
  if (selectedApiKeyId.value == null) return apiKeys.value.find((item) => item.key === apiKeyInput.value.trim()) || null
  return apiKeys.value.find((item) => item.id === selectedApiKeyId.value) || null
})

const isNanoBananaImageModel = computed(() => isNanoBananaModel(imageModel.value))
const isNanoBananaEditModel = computed(() => isNanoBananaModel(imageEditModel.value))

const getBase64Payload = (dataUrl: string) => {
  const marker = ';base64,'
  const index = dataUrl.indexOf(marker)
  return index >= 0 ? dataUrl.slice(index + marker.length) : dataUrl
}

const sleep = (ms: number) => new Promise((resolve) => window.setTimeout(resolve, ms))

const getNanoBananaFrontendTimeout = (imageSize: string) => {
  switch ((imageSize || '').toUpperCase()) {
    case '4K':
      return 10 * 60 * 1000
    case '2K':
      return 6 * 60 * 1000
    default:
      return 4 * 60 * 1000
  }
}

const getNanoBananaEstimatedDurationMs = (imageSize: string) => {
  return (imageSize || '').toUpperCase() === '4K' ? 2 * 60 * 1000 : 60 * 1000
}

const formatDuration = (ms: number) => {
  const totalSeconds = Math.max(1, Math.round(ms / 1000))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  if (minutes <= 0) return `${totalSeconds}秒`
  if (seconds === 0) return `${minutes}分`
  return `${minutes}分${seconds}秒`
}

const stopImageProcessingTimer = () => {
  if (imageProcessingTimer != null) {
    window.clearInterval(imageProcessingTimer)
    imageProcessingTimer = null
  }
}

const resetImageProcessingProgress = () => {
  stopImageProcessingTimer()
  imageProcessingProgress.value = 0
  imageProcessingStatus.value = ''
}

const setImageProcessingProgress = (progress: number, status?: string) => {
  imageProcessingProgress.value = Math.max(imageProcessingProgress.value, Math.min(100, Math.round(progress)))
  if (status) imageProcessingStatus.value = status
}

const startImageProcessingTimer = (estimatedMs: number) => {
  resetImageProcessingProgress()
  const startedAt = Date.now()
  imageProcessingStatus.value = '准备提交任务...'
  imageProcessingTimer = window.setInterval(() => {
    const elapsed = Date.now() - startedAt
    const estimatedProgress = Math.min(92, (elapsed / estimatedMs) * 100)
    setImageProcessingProgress(estimatedProgress, imageProcessingStatus.value || '处理中...')
  }, 250)
}

const finishImageProcessingProgress = async () => {
  stopImageProcessingTimer()
  const start = imageProcessingProgress.value
  const delta = Math.max(0, 100 - start)
  const steps = Math.max(1, Math.ceil(delta / 8))
  for (let index = 1; index <= steps; index += 1) {
    imageProcessingProgress.value = Math.min(100, Math.round(start + (delta * index) / steps))
    imageProcessingStatus.value = '即将完成...'
    await sleep(40)
  }
}

const readImageAsDataURL = (file: File) => new Promise<string>((resolve, reject) => {
  const reader = new FileReader()
  reader.onload = (event) => resolve((event.target?.result as string) || '')
  reader.onerror = () => reject(new Error('文件读取失败'))
  reader.readAsDataURL(file)
})

const validateImageFile = (file: File | null) => {
  if (!file) return false
  if (!file.type.startsWith('image/')) {
    appStore.showError('请上传图片文件')
    return false
  }
  if (file.size > 10 * 1024 * 1024) {
    appStore.showError('图片文件不能超过 10MB')
    return false
  }
  return true
}

const appendImageReferences = async (files: File[]) => {
  const validFiles = files.filter((file) => validateImageFile(file))
  if (validFiles.length === 0) return
  const previews = await Promise.all(validFiles.map((file) => readImageAsDataURL(file)))
  const nextItems = validFiles.map((file, index) => ({
    key: `${file.name}-${file.size}-${file.lastModified}-${Date.now()}-${index}`,
    file,
    preview: previews[index],
    base64: getBase64Payload(previews[index])
  }))
  imageReferenceItems.value = [...imageReferenceItems.value, ...nextItems]
}

const handleImageReferenceFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])
  if (files.length === 0) return
  try {
    await appendImageReferences(files)
  } catch (error: any) {
    appStore.showError(error.message || '参考图读取失败')
  } finally {
    target.value = ''
  }
}

const handleImageReferenceDrop = async (event: DragEvent) => {
  imageReferenceDragover.value = false
  const files = Array.from(event.dataTransfer?.files || [])
  if (files.length === 0) return
  try {
    await appendImageReferences(files)
  } catch (error: any) {
    appStore.showError(error.message || '参考图读取失败')
  }
}

const removeImageReference = (index: number) => {
  imageReferenceItems.value.splice(index, 1)
}

const clearImageReferencePreview = () => {
  imageReferenceItems.value = []
  imageReferenceDragover.value = false
  if (imageReferenceFileInput.value) {
    imageReferenceFileInput.value.value = ''
  }
}

const normalizeMultiplierValue = (value?: number | null) => {
  const normalized = typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : 1
  const rounded = Math.round(normalized * 10000) / 10000
  return Math.abs(rounded - 1) < 0.0001 ? 1 : rounded
}

const effectiveRate = computed(() => normalizeMultiplierValue(effectiveRateForGroup(activeApiKey.value?.group_id ?? selectedGroupId.value)))
const effectiveRateLabel = computed(() => t('modelTest.pricing.effectiveRate', { rate: effectiveRate.value.toFixed(2) }))
const pricedModelsCount = computed(() => Object.values(pricingMap.value).filter((item) => item.pricing_available).length)

const pricedModels = computed(() =>
  models.value.map((model) => {
    const pricing = pricingMap.value[model.id]
    const standardInputPrice = pricing?.input_price_per_1m || 0
    const standardOutputPrice = pricing?.output_price_per_1m || 0
    const imagePricePerImage = pricing?.image_price_per_image || 0
    const videoPricePerRequest = pricing?.video_price_per_request || 0
    const videoPricePerRequestHD = pricing?.video_price_per_request_hd || 0
    return {
      ...model,
      pricingAvailable: pricing?.pricing_available || false,
      standardInputPrice,
      standardOutputPrice,
      imagePricePerImage,
      videoPricePerRequest,
      videoPricePerRequestHD,
      actualInputPrice: standardInputPrice * effectiveRate.value,
      actualOutputPrice: standardOutputPrice * effectiveRate.value
    }
  })
)

const maskKey = (value: string) => (value.length <= 12 ? value : `${value.slice(0, 8)}...${value.slice(-4)}`)

const loadBootstrap = async () => {
  loadingBootstrap.value = true
  try {
    const [keyResult, availableGroups, rates] = await Promise.all([
      keysAPI.list(1, 1000),
      userGroupsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates()
    ])
    apiKeys.value = keyResult.items || []
    groups.value = availableGroups || []
    userGroupRates.value = rates || {}
    if (selectedGroupId.value == null && groups.value.length > 0) {
      selectedGroupId.value = groups.value[0].id
    }
  } catch (error) {
    console.error('Failed to load model test bootstrap data:', error)
    appStore.showError(t('modelTest.bootstrapFailed'))
  } finally {
    loadingBootstrap.value = false
  }
}

const applySelectedApiKey = () => {
  const matched = apiKeys.value.find((item) => item.id === selectedApiKeyId.value)
  if (!matched) return
  apiKeyInput.value = matched.key
  if (matched.group_id != null) {
    selectedGroupId.value = matched.group_id
  }
}

const fetchGatewayJSON = async (path: string, payload?: unknown) => {
  const response = await fetch(path, {
    method: payload ? 'POST' : 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKeyInput.value.trim()}`
    },
    body: payload ? JSON.stringify(payload) : undefined
  })
  const text = await response.text()
  const contentType = response.headers.get('content-type') || ''
  if (text && !contentType.includes('application/json')) {
    throw new Error(t('modelTest.models.nonJsonResponse'))
  }
  let parsed: any = null
  try {
    parsed = text ? JSON.parse(text) : null
  } catch {
    throw new Error(t('modelTest.models.nonJsonResponse'))
  }
  if (!response.ok) {
    throw new Error(parsed?.error?.message || parsed?.message || `HTTP ${response.status}`)
  }
  return parsed
}

const loadModels = async () => {
  if (!apiKeyInput.value.trim()) {
    appStore.showError(t('modelTest.keyPanel.directInputRequired'))
    return
  }
  loadingModels.value = true
  try {
    const data = await fetchGatewayJSON('/v1/models')
    models.value = Array.isArray(data?.data) ? data.data : []
    if (!chatModel.value && models.value.length > 0) {
      chatModel.value = models.value[0].id
    }
    if (!imageModel.value && imageModelOptions.value.length > 0) {
      imageModel.value = imageModelOptions.value[0].value
    }
    const pricing = await usageAPI.getModelPricingPreview(models.value.map((item) => item.id), apiKeyInput.value.trim())
    pricingMap.value = Object.fromEntries((pricing.models || []).map((item) => [item.model, item]))
  } catch (error: any) {
    appStore.showError(error.message || t('modelTest.models.fetchFailed'))
  } finally {
    loadingModels.value = false
  }
}

const selectModel = (modelId: string) => {
  chatModel.value = modelId
}

const generateApiKey = async () => {
  if (selectedGroupId.value == null) {
    appStore.showError(t('modelTest.keyPanel.groupRequired'))
    return
  }
  generatingKey.value = true
  try {
    const created = await keysAPI.create(`model-test-${Date.now()}`, selectedGroupId.value)
    apiKeys.value = [created, ...apiKeys.value]
    selectedApiKeyId.value = created.id
    apiKeyInput.value = created.key
    await copyToClipboard(created.key)
    appStore.showSuccess(t('modelTest.keyPanel.generatedAndCopied'))
    await loadModels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('modelTest.keyPanel.generateFailed'))
  } finally {
    generatingKey.value = false
  }
}

const clearChat = () => {
  chatMessages.value = []
  chatInput.value = ''
}

const scrollChatToBottom = async () => {
  await nextTick()
  if (chatScrollRef.value) {
    chatScrollRef.value.scrollTop = chatScrollRef.value.scrollHeight
  }
}

const sendMessage = async () => {
  const prompt = chatInput.value.trim()
  if (!prompt || !chatModel.value || !apiKeyInput.value.trim()) return
  chatMessages.value.push({ role: 'user', content: prompt })
  chatInput.value = ''
  await scrollChatToBottom()
  sending.value = true
  try {
    const input = chatMessages.value.map((message) => ({
      role: message.role,
      content: [{ type: 'input_text', text: message.content }]
    }))
    const result = await fetchGatewayJSON('/v1/responses', { model: chatModel.value, input })
    const reply =
      result?.output_text ||
      result?.output?.find?.((item: any) => item.type === 'message')?.content?.find?.((item: any) => item.type === 'output_text')?.text ||
      t('modelTest.chat.emptyReply')
    chatMessages.value.push({ role: 'assistant', content: reply })
    await scrollChatToBottom()
  } catch (error: any) {
    appStore.showError(error.message || t('modelTest.chat.failed'))
  } finally {
    sending.value = false
  }
}

const pollNanoBananaResult = async (taskID: string, imageSize: string) => {
	const startedAt = Date.now()
	const timeout = getNanoBananaFrontendTimeout(imageSize)
	while (Date.now() - startedAt < timeout) {
		const result = await fetchGatewayJSON('/v1/draw/result', { id: taskID }) as NanoBananaResultResponse
		if ((result?.code ?? 0) !== 0) {
			throw new Error(result?.msg || 'Nano Banana 获取结果失败')
		}
		const task = result?.data
		const status = (task?.status || '').toLowerCase()
		setImageProcessingProgress(task?.progress || 0, status === 'running' ? `处理中 ${task?.progress || 0}%` : '处理中...')
		if (status === 'succeeded') {
			return task || {}
		}
		if (status === 'failed') {
			throw new Error(task?.error || task?.failure_reason || 'Nano Banana 处理失败')
		}
		await sleep(2000)
	}
	throw new Error('Nano Banana 处理超时，请稍后去结果接口轮询')
}

const processNanoBananaImage = async (prompt: string) => {
	startImageProcessingTimer(getNanoBananaEstimatedDurationMs(nanoBananaImageSize.value))
	const createResult = await fetchGatewayJSON('/v1/draw/nano-banana', {
		model: imageModel.value,
		prompt,
		aspectRatio: nanoBananaAspectRatio.value,
		imageSize: nanoBananaImageSize.value,
		urls: imageReferenceItems.value.length > 0 ? imageReferenceItems.value.map((item) => item.base64) : undefined,
		webHook: '-1',
		shutProgress: true
	}) as NanoBananaCreateResponse
	if ((createResult?.code ?? 0) !== 0 || !createResult?.data?.id) {
		throw new Error(createResult?.msg || 'Nano Banana 提交任务失败')
	}
	imageProcessingStatus.value = '任务已提交，等待生成结果...'
	const task = await pollNanoBananaResult(createResult.data.id, nanoBananaImageSize.value)
	await finishImageProcessingProgress()
	const durationMs = task.start_time && task.end_time && task.end_time >= task.start_time
		? (task.end_time - task.start_time) * 1000
		: null
	return {
		urls: (task.results || []).map((item) => item.url || '').filter(Boolean),
		durationMs
	}
}

const generateImage = async () => {
  const prompt = imagePrompt.value.trim()
  if (!prompt || !imageModel.value || !apiKeyInput.value.trim()) return
  generatingImage.value = true
  const startedAt = Date.now()
  try {
    if (imageReferenceItems.value.length > 0 && !isNanoBananaImageModel.value) {
      throw new Error('参考图处理仅支持 Nano Banana 系列模型，请切换模型后再试')
    }
    if (isNanoBananaImageModel.value) {
      const taskResult = await processNanoBananaImage(prompt)
      generatedImages.value = taskResult.urls
      imageLastDurationMs.value = taskResult.durationMs || (Date.now() - startedAt)
    } else {
      const result = await fetchGatewayJSON('/v1/images/generations', {
        model: imageModel.value,
        prompt,
        n: 1,
        size: imageSize.value,
        response_format: 'url'
      }) as ImageGenerationResponse
      generatedImages.value = (result.data || []).map((item) => item.url || '').filter(Boolean)
    }
    if (generatedImages.value.length === 0) {
      throw new Error(t('modelTest.image.failed'))
    }
    // Add to history
    addToHistory({
      type: 'image',
      prompt,
      model: imageModel.value,
      size: isNanoBananaImageModel.value ? nanoBananaImageSize.value : imageSize.value,
      timestamp: Date.now()
    })
    if (!isNanoBananaImageModel.value) {
      imageLastDurationMs.value = Date.now() - startedAt
    }
  } catch (error: any) {
    resetImageProcessingProgress()
    appStore.showError(error.message || t('modelTest.image.failed'))
  } finally {
    if (!isNanoBananaImageModel.value) {
      resetImageProcessingProgress()
    }
    generatingImage.value = false
  }
}

const handleImageEditFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    imageEditFile.value = file
    const reader = new FileReader()
    reader.onload = (e) => {
      imageEditPreview.value = e.target?.result as string
    }
    reader.readAsDataURL(file)
  }
}

const editImage = async () => {
  const prompt = imageEditPrompt.value.trim()
  if (!prompt || !imageEditModel.value || !apiKeyInput.value.trim() || !imageEditFile.value) return
  editingImage.value = true
  try {
    const result = await editImageAPI({
      apiKey: apiKeyInput.value.trim(),
      model: imageEditModel.value,
      prompt,
      image: imageEditFile.value,
      n: 1,
      ...(isNanoBananaEditModel.value
        ? {
            aspect_ratio: nanoBananaEditAspectRatio.value,
            image_size: nanoBananaEditImageSize.value
          }
        : {
            size: imageEditSize.value
          }),
      response_format: 'url'
    })
    editedImages.value = (result.data || []).map((item) => item.url || '').filter(Boolean)
    if (editedImages.value.length === 0) {
      throw new Error('图片编辑失败')
    }
    // Add to history
    addToHistory({
      type: 'edit',
      prompt,
      model: imageEditModel.value,
      size: imageEditSize.value,
      timestamp: Date.now()
    })
    appStore.showSuccess('图片编辑成功')
  } catch (error: any) {
    appStore.showError(error.message || '图片编辑失败')
  } finally {
    editingImage.value = false
  }
}

const handleVideoReferenceImageUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    appStore.showError('请上传图片文件')
    return
  }

  // Validate file size (max 10MB)
  if (file.size > 10 * 1024 * 1024) {
    appStore.showError('图片文件不能超过 10MB')
    return
  }

  videoReferenceImageFile.value = file
  const reader = new FileReader()
  reader.onload = (e) => {
    videoReferenceImage.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const removeVideoReferenceImage = () => {
  videoReferenceImage.value = null
  videoReferenceImageFile.value = null
}

const generateVideo = async () => {
  const prompt = videoPrompt.value.trim()
  if (!prompt || !videoModel.value || !apiKeyInput.value.trim()) return
  generatingVideo.value = true
  try {
    const params: any = {
      apiKey: apiKeyInput.value.trim(),
      model: videoModel.value,
      prompt,
      size: videoSize.value,
      seconds: videoSeconds.value,
      quality: videoQuality.value
    }

    // Add reference image if uploaded
    if (videoReferenceImage.value) {
      params.image_reference = {
        image_url: videoReferenceImage.value
      }
    }

    const result = await generateVideoAPI(params)
    generatedVideos.value = (result.data || []).map((item) => item.url || '').filter(Boolean)
    if (generatedVideos.value.length === 0) {
      throw new Error('视频生成失败')
    }
    // Add to history
    addToHistory({
      type: 'video',
      prompt,
      model: videoModel.value,
      size: videoSize.value,
      quality: videoQuality.value,
      seconds: videoSeconds.value,
      timestamp: Date.now()
    })
    appStore.showSuccess('视频生成成功')
  } catch (error: any) {
    appStore.showError(error.message || '视频生成失败')
  } finally {
    generatingVideo.value = false
  }
}

const formatMultiplier = (value: number) => Number(value.toFixed(4)).toString()

const formatPrice = (value: number, available: boolean) => (available ? `$${value.toFixed(2)}` : '--')

const formatImageOrVideoPrice = (modelId: string, value: number, available: boolean) => {
  if (!available) return '--'
  if (isNanoBananaModel(modelId)) return `倍率 ${formatMultiplier(value)}`
  return formatPrice(value, true)
}

// Computed costs
const estimatedImageCost = computed(() => {
  if (!imageModel.value) return 0
  const pricing = pricingMap.value[imageModel.value]
  if (!pricing?.image_price_per_image) return 0
  if (isNanoBananaImageModel.value) return pricing.image_price_per_image
  return pricing.image_price_per_image * effectiveRate.value
})

const estimatedImageCostLabel = computed(() => {
  if (estimatedImageCost.value <= 0) return ''
  if (isNanoBananaImageModel.value) return `倍率：${formatMultiplier(estimatedImageCost.value)}`
  return `预计费用：$${estimatedImageCost.value.toFixed(4)}`
})

const estimatedImageDurationLabel = computed(() => {
  if (!isNanoBananaImageModel.value) return ''
  return `预计用时：约${formatDuration(getNanoBananaEstimatedDurationMs(nanoBananaImageSize.value))}`
})

const imageLastDurationLabel = computed(() => {
  if (!imageLastDurationMs.value) return ''
  return `上次用时：${formatDuration(imageLastDurationMs.value)}`
})

const estimatedEditCost = computed(() => {
  if (!imageEditModel.value) return 0
  const pricing = pricingMap.value[imageEditModel.value]
  if (!pricing?.image_price_per_image) return 0
  if (isNanoBananaEditModel.value) return pricing.image_price_per_image
  return pricing.image_price_per_image * effectiveRate.value
})

const estimatedEditCostLabel = computed(() => {
  if (estimatedEditCost.value <= 0) return ''
  if (isNanoBananaEditModel.value) return `倍率：${formatMultiplier(estimatedEditCost.value)}`
  return `预计费用：$${estimatedEditCost.value.toFixed(4)}`
})

const estimatedVideoCost = computed(() => {
  if (!videoModel.value) return 0
  const pricing = pricingMap.value[videoModel.value]
  if (!pricing) return 0
  const basePrice = videoQuality.value === 'high' 
    ? (pricing.video_price_per_request_hd || pricing.video_price_per_request || 0)
    : (pricing.video_price_per_request || 0)
  return basePrice * effectiveRate.value
})

// Helper functions for image/video operations
const appendToImagePrompt = (style: string) => {
  if (imagePrompt.value.trim()) {
    imagePrompt.value += `，${style}`
  } else {
    imagePrompt.value = style
  }
}

const appendToVideoPrompt = (movement: string) => {
  if (videoPrompt.value.trim()) {
    videoPrompt.value += `，${movement}`
  } else {
    videoPrompt.value = movement
  }
}

const downloadImage = async (url: string, filename: string) => {
  try {
    const response = await fetch(url)
    const blob = await response.blob()
    const blobUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = blobUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(blobUrl)
    appStore.showSuccess('下载成功')
  } catch (error) {
    appStore.showError('下载失败')
  }
}

const downloadVideo = async (url: string, filename: string) => {
  try {
    const response = await fetch(url)
    const blob = await response.blob()
    const blobUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = blobUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(blobUrl)
    appStore.showSuccess('下载成功')
  } catch (error) {
    appStore.showError('下载失败')
  }
}

const copyImageUrl = async (url: string) => {
  await copyToClipboard(url)
  appStore.showSuccess('图片链接已复制')
}

const copyVideoUrl = async (url: string) => {
  await copyToClipboard(url)
  appStore.showSuccess('视频链接已复制')
}

const removeGeneratedImage = (index: number) => {
  generatedImages.value.splice(index, 1)
  appStore.showSuccess('已删除')
}

const removeEditedImage = (index: number) => {
  editedImages.value.splice(index, 1)
  appStore.showSuccess('已删除')
}

const removeGeneratedVideo = (index: number) => {
  generatedVideos.value.splice(index, 1)
  appStore.showSuccess('已删除')
}

const clearImageEditPreview = () => {
  imageEditPreview.value = null
  imageEditFile.value = null
  if (imageEditFileInput.value) {
    imageEditFileInput.value.value = ''
  }
}

const handleImageEditDrop = (event: DragEvent) => {
  imageEditDragover.value = false
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (file.type.startsWith('image/')) {
      imageEditFile.value = file
      const reader = new FileReader()
      reader.onload = (e) => {
        imageEditPreview.value = e.target?.result as string
      }
      reader.readAsDataURL(file)
    } else {
      appStore.showError('请上传图片文件')
    }
  }
}

watch(selectedGroupId, () => {
  if (models.value.length > 0) {
    void loadModels()
  }
})

watch(imageModelOptions, (options) => {
  if (!imageModel.value && options.length > 0) {
    imageModel.value = options[0].value
  }
})

watch(imageModel, (model) => {
  if (!isNanoBananaModel(model)) return
  const supportedSizes = getNanoBananaSupportedImageSizes(model)
  if (!supportedSizes.includes(nanoBananaImageSize.value)) {
    nanoBananaImageSize.value = getNanoBananaDefaultImageSize(model)
  }
  const supportedRatios = getNanoBananaAspectRatioOptions(model)
  if (!supportedRatios.includes(nanoBananaAspectRatio.value)) {
    nanoBananaAspectRatio.value = 'auto'
  }
})

watch(isNanoBananaImageModel, (enabled) => {
  if (!enabled) {
    clearImageReferencePreview()
  }
})

watch(imageEditModelOptions, (options) => {
  if (!imageEditModel.value && options.length > 0) {
    imageEditModel.value = options[0].value
  }
})

watch(imageEditModel, (model) => {
  if (!isNanoBananaModel(model)) return
  const supportedSizes = getNanoBananaSupportedImageSizes(model)
  if (!supportedSizes.includes(nanoBananaEditImageSize.value)) {
    nanoBananaEditImageSize.value = getNanoBananaDefaultImageSize(model)
  }
  const supportedRatios = getNanoBananaAspectRatioOptions(model)
  if (!supportedRatios.includes(nanoBananaEditAspectRatio.value)) {
    nanoBananaEditAspectRatio.value = 'auto'
  }
})

watch(videoModelOptions, (options) => {
  if (!videoModel.value && options.length > 0) {
    videoModel.value = options[0].value
  }
})

onMounted(() => {
  void loadBootstrap()
  loadHistory()
})
</script>
