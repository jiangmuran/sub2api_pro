<template>
  <AppLayout>
    <section class="mx-auto max-w-4xl space-y-8">
      <header>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('legal.termsTitle') }} / {{ t('legal.termsTitleEn') }}
        </h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('legal.lastUpdated') }}: 2025-01-01
        </p>
      </header>

      <article class="space-y-6 text-sm leading-relaxed text-gray-700 dark:text-dark-100">
        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            一、适用范围 / Scope
          </h2>
          <p class="mt-2">
            本《用户协议》（以下简称“本协议”）适用于您使用本服务端点及相关网站、客户端、接口、文档
            （统称“本服务”）的行为。无论您是以个人还是组织身份访问，本协议均适用。
          </p>
          <p class="mt-2">
            This Terms of Service (the "Terms") govern your access to and use of this service, including all
            related websites, APIs, SDKs, documentation and tools (collectively, the "Service"), whether you
            are acting as an individual or on behalf of an organization.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            二、接受条款 / Acceptance
          </h2>
          <p class="mt-2">
            您在注册、登录或实际使用本服务前，应当仔细阅读并充分理解本协议全部内容。一旦您勾选或点击
            “同意”“注册”“登录”或实际使用本服务，即视为您已阅读并同意受本协议全部条款的约束。
          </p>
          <p class="mt-2">
            By clicking "Accept", "Sign up", "Log in" or by accessing or using the Service, you acknowledge
            that you have read, understood and agree to be bound by these Terms. If you do not agree, you must
            not use the Service.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            三、账户与安全 / Account & Security
          </h2>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>您应对账号及凭据（包括 API Key、Token 等）的保管负责，不得出租、出借、转让或共享。</li>
            <li>如发现账号或凭据存在泄露、被盗用等安全风险，应立即修改密码并通知服务提供方。</li>
            <li>因您保管不当、操作不当或第三方原因导致的损失，由您自行承担。</li>
          </ul>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>You are responsible for maintaining the confidentiality of your account credentials, including any API keys or tokens.</li>
            <li>You must promptly notify the operator of any suspected unauthorized use or security incident.</li>
            <li>The Service is provided on an "as is" and "as available" basis, and you bear sole responsibility for your integration and data security on your side.</li>
          </ul>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            四、合理使用与禁止行为 / Acceptable Use
          </h2>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>不得利用本服务从事任何违反适用法律法规的行为。</li>
            <li>不得用于生成、传播或辅助传播违法、有害、侵权、欺诈或误导性内容。</li>
            <li>不得试图绕过计费、配额、身份验证或安全机制。</li>
            <li>不得对本服务进行恶意攻击、逆向工程或未授权的自动化抓取。</li>
          </ul>
          <ul class="mt-2 list-disc space-y-1 pl-5">
            <li>You must not use the Service in any way that violates applicable laws or regulations.</li>
            <li>You must not use the Service to generate or distribute illegal, harmful, infringing, fraudulent or misleading content.</li>
            <li>You must not attempt to bypass billing, rate limits, authentication or security controls.</li>
            <li>You must not attack, reverse engineer or scrape the Service in an unauthorized manner.</li>
          </ul>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            五、第三方服务与模型 / Third-Party Services & Models
          </h2>
          <p class="mt-2">
            本服务可能通过网关方式转发至第三方模型或 API（包括但不限于 OpenAI、Anthropic 等）。第三方服
            务的具体能力、可用性、价格、合规要求等由该第三方决定并可能随时变化，您应自行阅读并遵守相应
            第三方的服务条款和政策。
          </p>
          <p class="mt-2">
            The Service may act as a gateway or proxy to third-party APIs or models (including without limitation
            OpenAI, Anthropic and others). Such third-party services are governed solely by their own terms and
            policies, which may change at any time. You are responsible for reviewing and complying with those
            third-party terms.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            六、免责声明与责任限制 / Disclaimer & Limitation of Liability
          </h2>
          <p class="mt-2">
            本服务按“现状”和“可用”基础提供，不对任何明示或默示的保证作出承诺，包括但不限于适销性、特
            定用途适用性、持续可用性或无错误等。AI 生成内容具有不确定性，仅供参考，不构成任何法律、医
            疗、投资等专业意见。
          </p>
          <p class="mt-2">
            在适用法律允许的最大范围内，本服务提供方对因您使用或无法使用本服务而产生的任何直接、间接、
            附带、特殊、衍生或惩罚性损害不承担责任，即使相关方已被告知可能发生此类损害。
          </p>
          <p class="mt-2">
            THE SERVICE IS PROVIDED ON AN "AS IS" AND "AS AVAILABLE" BASIS. TO THE MAXIMUM EXTENT PERMITTED BY
            LAW, THE OPERATOR DISCLAIMS ALL WARRANTIES OF ANY KIND, WHETHER EXPRESS OR IMPLIED, AND SHALL NOT BE
            LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL OR PUNITIVE DAMAGES ARISING OUT OF OR IN
            CONNECTION WITH YOUR USE OF THE SERVICE.
          </p>
        </section>

        <section>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            七、变更与终止 / Changes & Termination
          </h2>
          <p class="mt-2">
            为适应业务发展、法律法规或监管要求，本服务有权对本协议进行调整并在网站合理位置公示。若您在
            协议更新生效后继续使用本服务，即视为接受更新后的条款。
          </p>
          <p class="mt-2">
            We may modify these Terms from time to time. Material changes will be announced in a reasonable manner
            (for example via this website or dashboard). Your continued use of the Service after such changes take
            effect constitutes your acceptance of the updated Terms.
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
