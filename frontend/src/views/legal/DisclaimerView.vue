<template>
  <AppLayout>
    <section class="mx-auto max-w-4xl space-y-8">
      <header>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('legal.disclaimerTitle') }} / {{ t('legal.disclaimerTitleEn') }}
        </h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('legal.lastUpdated') }}: 2026-03-11
        </p>
      </header>

      <article class="space-y-8 text-sm leading-relaxed text-gray-700 dark:text-dark-100">
        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            一、地理限制与访问政策（严格遵守 OFAC 及本地法规）
          </h2>
          <ol class="list-decimal space-y-2 pl-5">
            <li>
              禁止中国大陆地区访问：我们严格遵守相关国家和地区的法律法规。明确禁止位于中华人民共和国大陆地区（包括香港、澳门特别行政区及台湾地区）的任何个人或实体访问、使用本服务。
            </li>
            <li>
              用户合规承诺：在访问或使用本服务前，您必须确保您所在的地理位置不在上述禁止区域内。如果您通过虚拟私人网络、代理服务器或其他手段绕过地理限制从禁止区域接入服务，您将被视为违反本协议，由此产生的一切后果及法律责任由您自行承担，本站不承担任何责任。
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            二、服务性质与来源说明（开源与非盈利）
          </h2>
          <ol class="list-decimal space-y-2 pl-5">
            <li>
              非官方与中立性：本服务仅为技术爱好者提供的开源项目演示或公益性质的中转代理，旨在促进技术交流。我们并非 OpenAI 或任何上游 API 提供商的官方合作伙伴、授权代理或附属机构。
            </li>
            <li>
              上游服务免责声明：本服务依赖于第三方上游 API 接口。我们不对上游服务的稳定性、准确性、合法性及可用性作任何明示或暗示的保证。上游服务的商标、版权及相关权利均归其原作者所有。
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            三、知识产权与账户责任
          </h2>
          <ol class="list-decimal space-y-2 pl-5">
            <li>
              账户合规性：本服务所使用的上游账户均由独立第三方提供。我们作为中转服务提供方，对上游账户的具体注册方式、支付手段及资金来源不知情，亦不参与其注册与维护过程。
            </li>
            <li>
              免责条款：您理解并同意，若因上游账户涉及违规（如使用虚拟信用卡、盗刷等行为）导致服务中断、API 密钥失效或数据丢失，属于不可抗力或第三方原因，本站不承担由此产生的直接或间接损失。信用卡相关（盗刷，诈骗）本站完全没有参与与知情，责任与本站无关。
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            四、信用卡与支付政策（零容忍欺诈）
          </h2>
          <ol class="list-decimal space-y-2 pl-5">
            <li>
              反欺诈声明：本站明确反对任何形式的支付欺诈、信用卡盗刷及非法洗钱活动。我们未参与上游账户的支付环节，亦未从中获取任何非法利益。
            </li>
            <li>
              配合调查：若司法机关或监管机构依法要求提供相关日志以协助调查上游欺诈行为，我们将依法配合，但不承担由此引发的连带责任。
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            五、开源与免费使用风险
          </h2>
          <ol class="list-decimal space-y-2 pl-5">
            <li>
              免费即无担保：本服务为免费提供，不收取任何费用。免费版本可能存在性能波动、访问延迟或功能限制。您需自行承担使用本服务带来的全部风险，包括但不限于数据泄露、服务中断等。
            </li>
            <li>
              数据安全：我们承诺绝对不会主动窥探、存储或出售您的传输数据。但由于中转架构的特性，建议您不要通过本服务传输高度敏感的个人隐私或商业机密。
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            六、最终解释权
          </h2>
          <p>
            本声明的最终解释权归本站所有。如本声明中的任何条款与现行法律相抵触，该条款将按法律规定重新解释，但不影响其他条款的效力。
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
