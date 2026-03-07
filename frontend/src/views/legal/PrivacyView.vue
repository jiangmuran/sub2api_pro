<template>
  <AppLayout>
    <section class="mx-auto max-w-4xl space-y-8">
      <header>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('legal.privacyTitle') }} / {{ t('legal.privacyTitleEn') }}
        </h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('legal.lastUpdated') }}: 2025-01-01
        </p>
      </header>

      <article class="space-y-6 text-sm leading-relaxed text-gray-700 dark:text-dark-100">
        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            一、我们收集的信息 / Information We Collect
          </h2>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>账号信息：邮箱、用户名、密码哈希等。</li>
            <li>使用数据：请求时间、模型名称、基础用量统计（如 token 数量），用于计费与运维分析。</li>
            <li>技术信息：IP 地址、浏览器/客户端信息、错误日志，用于安全审计与问题排查。</li>
          </ul>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>Account data such as email address, username and hashed passwords.</li>
            <li>Usage data such as timestamps, model names and aggregate usage metrics for billing and operations.</li>
            <li>Technical data such as IP address, user agent and error logs for security and troubleshooting.</li>
          </ul>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            二、数据用途 / How We Use Data
          </h2>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>提供与维护服务，包括认证、计费、风控与性能优化。</li>
            <li>改进产品体验，例如根据整体使用情况优化容量规划与路由策略。</li>
            <li>履行法律义务或合规要求（如日志留存、协助调查等）。</li>
          </ul>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>To provide, operate and improve the Service, including authentication, billing, rate limiting and routing.</li>
            <li>To analyze aggregated usage patterns for capacity planning and reliability.</li>
            <li>To comply with legal obligations, enforcement requests or applicable regulations.</li>
          </ul>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            三、数据存储与保留 / Storage & Retention
          </h2>
          <p class="mt-2">
            我们会在完成目的所必需的期间内保存您的数据，并在法律要求或争议解决所需的范围内适当延长保留期。
            对于日志、调用记录等技术数据，我们会采取合理的访问控制与安全措施，并按业务需要定期清理或匿名化。
          </p>
          <p class="mt-2">
            We retain personal data for as long as necessary to fulfill the purposes outlined in this Policy, and for
            a reasonable additional period where required by law or to resolve disputes. Technical logs and request
            metadata are protected by access controls and may be periodically deleted or anonymized.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            四、数据共享与第三方 / Sharing & Third Parties
          </h2>
          <p class="mt-2">
            我们不会向不相关的第三方出售您的个人数据。为提供模型能力和基础设施，本服务可能会将必要的请求
            数据转发给第三方模型服务商或基础设施提供商（例如云服务商），并尽量在协议与技术上限制其用途。
          </p>
          <p class="mt-2">
            We do not sell your personal data to unrelated third parties. To deliver the Service, we may pass necessary
            data to model providers or infrastructure vendors, subject to contractual and technical safeguards where
            reasonably feasible.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            五、您的权利 / Your Rights
          </h2>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>有权访问、更新或删除您的账户信息（在技术和法律允许范围内）。</li>
            <li>有权选择停止使用本服务，并请求关闭账号。关闭账号后，我们可能仍需在合理期限内保留部分日志用于合规与安全审计。</li>
          </ul>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>You may request access, correction or deletion of your account data where technically and legally feasible.</li>
            <li>You may stop using the Service at any time and request account closure. Certain records may be retained for a limited period for compliance and security.</li>
          </ul>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            六、安全措施 / Security
          </h2>
          <p class="mt-2">
            我们会采取合理的技术和组织措施保护您的数据安全，但任何网络服务都无法保证绝对安全。您应当在
            自己一侧采取必要的安全措施（例如妥善保管 API Key、定期轮换凭据、限制访问范围等）。
          </p>
          <p class="mt-2">
            We implement reasonable technical and organizational measures to protect your data, but no system is
            perfectly secure. You remain responsible for your own environment, including protecting API keys and
            access credentials.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            七、政策更新 / Updates
          </h2>
          <p class="mt-2">
            我们可能会不时更新本隐私政策，并在网站上以合理方式提示。重大变更在生效前会通过公告等方式提醒。
            如您在更新生效后继续使用本服务，即视为接受更新后的政策。
          </p>
          <p class="mt-2">
            We may revise this Privacy Policy from time to time. Material changes will be announced in a reasonable
            manner. Your continued use of the Service after such changes take effect constitutes acceptance of the
            updated Policy.
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
