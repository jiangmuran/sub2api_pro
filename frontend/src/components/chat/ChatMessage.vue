<template>
  <div 
    :class="[
      'flex gap-3 mb-4',
      message.role === 'user' ? 'flex-row-reverse' : 'flex-row'
    ]"
  >
    <!-- Avatar -->
    <div class="flex-shrink-0">
      <div 
        :class="[
          'w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold',
          message.role === 'user' 
            ? 'bg-gradient-to-br from-cyan-500 to-teal-600 text-white' 
            : 'bg-gradient-to-br from-teal-500 to-cyan-600 text-white'
        ]"
      >
        {{ message.role === 'user' ? '你' : 'AI' }}
      </div>
    </div>

    <!-- Message Content -->
    <div 
      :class="[
        'flex-1 max-w-[85%]',
        message.role === 'user' ? 'items-end' : 'items-start'
      ]"
    >
      <!-- Message Bubble -->
      <div 
        :class="[
          'rounded-2xl px-4 py-3 shadow-lg',
          message.role === 'user'
            ? 'bg-gradient-to-br from-cyan-500 to-teal-600 text-white ml-auto'
            : 'bg-slate-800 text-slate-100'
        ]"
      >
        <!-- Streaming Text with Cursor -->
        <div v-if="isStreaming" class="flex items-center gap-1">
          <span class="whitespace-pre-wrap">{{ message.content }}</span>
          <span class="inline-block w-2 h-5 bg-current animate-pulse"></span>
        </div>

        <!-- Rendered Markdown -->
        <div 
          v-else-if="message.content"
          class="prose prose-sm prose-invert max-w-none"
          v-html="renderedContent"
        />

        <!-- Empty State -->
        <div v-else class="text-slate-400 italic">
          Empty message
        </div>
      </div>

      <!-- Meta Information -->
      <div 
        :class="[
          'flex items-center gap-2 mt-1 px-2 text-xs text-slate-500',
          message.role === 'user' ? 'justify-end' : 'justify-start'
        ]"
      >
        <span>{{ formatTime(message.timestamp) }}</span>
      </div>

      <!-- Action Buttons -->
      <div 
        v-if="!isStreaming"
        :class="[
          'flex gap-1 mt-2',
          message.role === 'user' ? 'justify-end' : 'justify-start'
        ]"
      >
        <button
          @click="copyMessage"
          class="p-1.5 rounded-lg hover:bg-slate-700 text-slate-400 hover:text-slate-200 transition-colors"
          title="Copy"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>

        <button
          v-if="message.role === 'assistant'"
          @click="$emit('regenerate')"
          class="p-1.5 rounded-lg hover:bg-slate-700 text-slate-400 hover:text-slate-200 transition-colors"
          title="Regenerate"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'
import type { Message } from '@/types/chat'

const props = defineProps<{
  message: Message
  isStreaming?: boolean
}>()

defineEmits<{
  regenerate: []
}>()

// Configure marked for code highlighting
marked.use({
  breaks: true,
  gfm: true,
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      if (lang && hljs.getLanguage(lang)) {
        try {
          const highlighted = hljs.highlight(text, { language: lang }).value
          return `<pre><code class="hljs ${lang}">${highlighted}</code></pre>`
        } catch (err) {
          console.warn('Highlight error:', err)
        }
      }
      const highlighted = hljs.highlightAuto(text).value
      return `<pre><code class="hljs">${highlighted}</code></pre>`
    }
  }
})

const renderedContent = computed(() => {
  if (!props.message.content) return ''
  try {
    const html = marked.parse(props.message.content) as string
    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: [
        'p', 'br', 'strong', 'em', 'u', 's', 'code', 'pre',
        'a', 'ul', 'ol', 'li', 'blockquote', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
        'table', 'thead', 'tbody', 'tr', 'th', 'td', 'span', 'div'
      ],
      ALLOWED_ATTR: ['href', 'class', 'id', 'target', 'rel']
    })
  } catch (err) {
    console.error('Markdown parsing error:', err)
    return DOMPurify.sanitize(props.message.content)
  }
})

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const copyMessage = async () => {
  try {
    await navigator.clipboard.writeText(props.message.content)
    // TODO: Show toast notification
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>

<style scoped>
/* Markdown prose styles */
.prose :deep(pre) {
  @apply bg-slate-900 rounded-lg p-4 overflow-x-auto my-3;
}

.prose :deep(code) {
  @apply bg-slate-700 px-1.5 py-0.5 rounded text-sm font-mono;
}

.prose :deep(pre code) {
  @apply bg-transparent p-0;
}

.prose :deep(table) {
  @apply w-full border-collapse my-4;
}

.prose :deep(th),
.prose :deep(td) {
  @apply border border-slate-600 px-3 py-2;
}

.prose :deep(th) {
  @apply bg-slate-700 font-semibold;
}

.prose :deep(a) {
  @apply text-teal-400 hover:text-teal-300 underline;
}

.prose :deep(blockquote) {
  @apply border-l-4 border-teal-500 pl-4 italic my-4;
}

.prose :deep(ul),
.prose :deep(ol) {
  @apply my-3 pl-6;
}

.prose :deep(li) {
  @apply my-1;
}

.prose :deep(h1),
.prose :deep(h2),
.prose :deep(h3),
.prose :deep(h4),
.prose :deep(h5),
.prose :deep(h6) {
  @apply font-bold mt-6 mb-3;
}

.prose :deep(h1) { @apply text-2xl; }
.prose :deep(h2) { @apply text-xl; }
.prose :deep(h3) { @apply text-lg; }
</style>
