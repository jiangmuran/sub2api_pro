/**
 * Chat interface type definitions
 */

export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  tokens?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
  cost?: number
  model?: string
}

export interface Conversation {
  id: string
  title: string
  model: string
  messages: Message[]
  createdAt: number
  updatedAt: number
  totalCost: number
  totalTokens: number
}

export interface Model {
  id: string
  name: string
  context_length?: number
  pricing?: {
    prompt: number // per million tokens
    completion: number // per million tokens
  }
}

export interface GeneratedImage {
  id: string
  prompt: string
  url: string
  dataUrl?: string // cached as data URL
  size: string
  model: string
  timestamp: number
  cost: number
  revisedPrompt?: string
}

export interface ChatCompletionRequest {
  model: string
  messages: Array<{
    role: string
    content: string
  }>
  stream?: boolean
  max_tokens?: number
  temperature?: number
}

export interface ChatCompletionChunk {
  id: string
  object: string
  created: number
  model: string
  choices: Array<{
    index: number
    delta: {
      role?: string
      content?: string
    }
    finish_reason: string | null
  }>
}

export interface ChatCompletionResponse {
  id: string
  object: string
  created: number
  model: string
  choices: Array<{
    index: number
    message: {
      role: string
      content: string
    }
    finish_reason: string
  }>
  usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

export interface ImageGenerationRequest {
  model: string
  prompt: string
  size?: string
  n?: number
  quality?: string
}

export interface ImageGenerationResponse {
  created: number
  data: Array<{
    url: string
    revised_prompt?: string
  }>
}

export interface ModelPricing {
  [modelId: string]: {
    prompt: number
    completion: number
  }
}
