<template>
  <AppLayout>
    <section class="mx-auto max-w-4xl space-y-8">
      <header>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('legal.disclaimerTitle') }} / {{ t('legal.disclaimerTitleEn') }}
        </h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('legal.lastUpdated') }}: 2025-01-01
        </p>
      </header>

      <article class="space-y-6 text-sm leading-relaxed text-gray-700 dark:text-dark-100">
        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            一、AI 内容性质说明 / Nature of AI-Generated Content
          </h2>
          <p class="mt-2">
            本服务可能通过第三方模型生成或转发内容，该等内容基于统计与概率推断，并不保证真实、准确、完
            整或适用于任何特定目的。您不应将 AI 生成内容视为专业意见（包括但不限于法律、医疗、投资、财
            务等），也不应据此做出重大决策。
          </p>
          <p class="mt-2">
            The Service may generate or proxy AI-generated content produced by third-party models. Such content is
            inherently probabilistic and may be inaccurate, incomplete or inappropriate for your specific use case.
            It does not constitute professional advice of any kind.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            二、合规与使用责任 / Compliance & Responsibility
          </h2>
          <p class="mt-2">
            您应对自己使用本服务的方式、输入与输出内容承担全部责任，并确保符合适用法律法规及所在行业的监
            管要求。对于您或您的下游用户使用本服务或依赖本服务所产生的任何直接或间接后果，本服务提供方
            在法律允许范围内不承担责任。
          </p>
          <p class="mt-2">
            You are solely responsible for how you use the Service and for any content you or your end users input,
            generate or distribute. You must ensure your use complies with all applicable laws, regulations and
            policies. To the maximum extent permitted by law, the operator is not liable for any consequences arising
            from your use of or reliance on the Service.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            三、不可控因素 / Uncontrollable Factors
          </h2>
          <p class="mt-2">
            由于互联网环境的不确定性，服务可能受到网络、第三方服务商、基础设施故障等因素影响而中断、延
            迟或出现错误响应。我们会尽商业上合理的努力维持服务稳定，但不对任何特定可用性或响应时间作出
            保证。
          </p>
          <p class="mt-2">
            The availability and performance of the Service may be affected by network conditions, third-party
            outages and other factors beyond our reasonable control. We do not guarantee uninterrupted operation,
            specific uptime or latency.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            四、风险自担 / Use at Your Own Risk
          </h2>
          <p class="mt-2">
            您理解并同意，使用本服务存在技术与合规风险，包括但不限于：数据误判、模型偏差、内容不适当、
            第三方服务变更等。您应在充分评估风险后决定是否以及如何使用本服务，并为自己的选择负责。
          </p>
          <p class="mt-2">
            You acknowledge that use of the Service involves technical and compliance risks, including potential
            misclassification, bias, inappropriate content and changes to third-party services. You are responsible
            for evaluating these risks and determining whether and how to use the Service.
          </p>
        </section>
      </article>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
</script>
