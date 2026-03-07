<template>
  <footer class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
    <span>{{ t('legal.footerPrefix') }}</span>
    <router-link
      to="/legal/terms"
      class="ml-1 text-primary-600 hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
    >
      {{ t('legal.termsLink') }}
    </router-link>
    <span> · </span>
    <router-link
      to="/legal/privacy"
      class="text-primary-600 hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
    >
      {{ t('legal.privacyLink') }}
    </router-link>
    <span> · </span>
    <router-link
      to="/legal/disclaimer"
      class="text-primary-600 hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
    >
      {{ t('legal.disclaimerLink') }}
    </router-link>
  </footer>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
