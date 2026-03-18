/**
 * Chat storage composable
 * Uses IndexedDB for storing conversations, messages, and images
 */

import { ref, computed } from 'vue'
import type { Conversation, Message, GeneratedImage } from '@/types/chat'

// IndexedDB database name and version
const DB_NAME = 'sub2api-chat'
const DB_VERSION = 1

// Stores
const STORE_CONVERSATIONS = 'conversations'
const STORE_MESSAGES = 'messages'
const STORE_IMAGES = 'images'

interface MessageWithConversation extends Message {
  conversationId: string
}

let dbInstance: IDBDatabase | null = null

/**
 * Initialize IndexedDB
 */
const initDB = (): Promise<IDBDatabase> => {
  return new Promise((resolve, reject) => {
    if (dbInstance) {
      resolve(dbInstance)
      return
    }

    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onerror = () => reject(request.error)
    request.onsuccess = () => {
      dbInstance = request.result
      resolve(request.result)
    }

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result

      // Conversations store
      if (!db.objectStoreNames.contains(STORE_CONVERSATIONS)) {
        const convStore = db.createObjectStore(STORE_CONVERSATIONS, { keyPath: 'id' })
        convStore.createIndex('updatedAt', 'updatedAt', { unique: false })
      }

      // Messages store
      if (!db.objectStoreNames.contains(STORE_MESSAGES)) {
        const msgStore = db.createObjectStore(STORE_MESSAGES, { keyPath: 'id' })
        msgStore.createIndex('conversationId', 'conversationId', { unique: false })
        msgStore.createIndex('timestamp', 'timestamp', { unique: false })
      }

      // Images store
      if (!db.objectStoreNames.contains(STORE_IMAGES)) {
        const imgStore = db.createObjectStore(STORE_IMAGES, { keyPath: 'id' })
        imgStore.createIndex('timestamp', 'timestamp', { unique: false })
      }
    }
  })
}

export function useChatStorage() {
  const conversations = ref<Conversation[]>([])
  const currentConversationId = ref<string | null>(null)
  const images = ref<GeneratedImage[]>([])

  const currentConversation = computed(() => {
    return conversations.value.find((c) => c.id === currentConversationId.value)
  })

  /**
   * Load all conversations
   */
  const loadConversations = async () => {
    const db = await initDB()
    const transaction = db.transaction(STORE_CONVERSATIONS, 'readonly')
    const store = transaction.objectStore(STORE_CONVERSATIONS)
    const index = store.index('updatedAt')
    const request = index.openCursor(null, 'prev') // Most recent first

    const results: Conversation[] = []

    return new Promise<Conversation[]>((resolve, reject) => {
      request.onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) {
          results.push(cursor.value)
          cursor.continue()
        } else {
          conversations.value = results
          resolve(results)
        }
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Save or update a conversation
   */
  const saveConversation = async (conversation: Conversation) => {
    const db = await initDB()
    const transaction = db.transaction(STORE_CONVERSATIONS, 'readwrite')
    const store = transaction.objectStore(STORE_CONVERSATIONS)

    // Create a plain object without any non-serializable properties
    // Use JSON.parse(JSON.stringify()) to deep clone and remove all non-serializable data
    const plainConversation = JSON.parse(JSON.stringify({
      id: conversation.id,
      title: conversation.title,
      model: conversation.model,
      messages: conversation.messages || [],
      createdAt: conversation.createdAt,
      updatedAt: conversation.updatedAt,
      totalCost: conversation.totalCost || 0,
      totalTokens: conversation.totalTokens || 0
    }))

    return new Promise<void>((resolve, reject) => {
      const request = store.put(plainConversation)
      request.onsuccess = () => {
        // Update local state
        const index = conversations.value.findIndex((c) => c.id === conversation.id)
        if (index >= 0) {
          conversations.value[index] = plainConversation
        } else {
          conversations.value.unshift(plainConversation)
        }
        resolve()
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Delete a conversation and its messages
   */
  const deleteConversation = async (conversationId: string) => {
    const db = await initDB()

    // Delete conversation
    const convTransaction = db.transaction(STORE_CONVERSATIONS, 'readwrite')
    const convStore = convTransaction.objectStore(STORE_CONVERSATIONS)
    await new Promise<void>((resolve, reject) => {
      const request = convStore.delete(conversationId)
      request.onsuccess = () => resolve()
      request.onerror = () => reject(request.error)
    })

    // Delete all messages for this conversation
    const msgTransaction = db.transaction(STORE_MESSAGES, 'readwrite')
    const msgStore = msgTransaction.objectStore(STORE_MESSAGES)
    const index = msgStore.index('conversationId')
    const range = IDBKeyRange.only(conversationId)
    const request = index.openCursor(range)

    await new Promise<void>((resolve, reject) => {
      request.onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) {
          cursor.delete()
          cursor.continue()
        } else {
          resolve()
        }
      }
      request.onerror = () => reject(request.error)
    })

    // Update local state
    conversations.value = conversations.value.filter((c) => c.id !== conversationId)
    if (currentConversationId.value === conversationId) {
      currentConversationId.value = null
    }
  }

  /**
   * Load messages for a conversation
   */
  const loadMessages = async (conversationId: string): Promise<Message[]> => {
    const db = await initDB()
    const transaction = db.transaction(STORE_MESSAGES, 'readonly')
    const store = transaction.objectStore(STORE_MESSAGES)
    const index = store.index('conversationId')
    const range = IDBKeyRange.only(conversationId)
    const request = index.openCursor(range)

    const messages: Message[] = []

    return new Promise((resolve, reject) => {
      request.onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) {
          const msgWithConv = cursor.value as MessageWithConversation
          const { conversationId: _, ...message } = msgWithConv
          messages.push(message)
          cursor.continue()
        } else {
          // Sort by timestamp
          messages.sort((a, b) => a.timestamp - b.timestamp)
          resolve(messages)
        }
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Save a message
   */
  const saveMessage = async (conversationId: string, message: Message) => {
    const db = await initDB()
    const transaction = db.transaction(STORE_MESSAGES, 'readwrite')
    const store = transaction.objectStore(STORE_MESSAGES)

    // Create plain object to ensure serializability
    const plainMessage = JSON.parse(JSON.stringify({
      id: message.id,
      role: message.role,
      content: message.content,
      timestamp: message.timestamp,
      tokens: message.tokens,
      cost: message.cost,
      model: message.model,
      conversationId
    }))

    return new Promise<void>((resolve, reject) => {
      const request = store.put(plainMessage)
      request.onsuccess = () => resolve()
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Load all images
   */
  const loadImages = async () => {
    const db = await initDB()
    const transaction = db.transaction(STORE_IMAGES, 'readonly')
    const store = transaction.objectStore(STORE_IMAGES)
    const index = store.index('timestamp')
    const request = index.openCursor(null, 'prev') // Most recent first

    const results: GeneratedImage[] = []

    return new Promise<GeneratedImage[]>((resolve, reject) => {
      request.onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result
        if (cursor) {
          results.push(cursor.value)
          cursor.continue()
        } else {
          images.value = results
          resolve(results)
        }
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Save an image
   */
  const saveImage = async (image: GeneratedImage) => {
    const db = await initDB()
    const transaction = db.transaction(STORE_IMAGES, 'readwrite')
    const store = transaction.objectStore(STORE_IMAGES)

    // Create plain object to ensure serializability
    const plainImage = JSON.parse(JSON.stringify(image))

    return new Promise<void>((resolve, reject) => {
      const request = store.put(plainImage)
      request.onsuccess = () => {
        images.value.unshift(plainImage)
        resolve()
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Delete an image
   */
  const deleteImage = async (imageId: string) => {
    const db = await initDB()
    const transaction = db.transaction(STORE_IMAGES, 'readwrite')
    const store = transaction.objectStore(STORE_IMAGES)

    return new Promise<void>((resolve, reject) => {
      const request = store.delete(imageId)
      request.onsuccess = () => {
        images.value = images.value.filter((img) => img.id !== imageId)
        resolve()
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Clear all images
   */
  const clearAllImages = async () => {
    const db = await initDB()
    const transaction = db.transaction(STORE_IMAGES, 'readwrite')
    const store = transaction.objectStore(STORE_IMAGES)

    return new Promise<void>((resolve, reject) => {
      const request = store.clear()
      request.onsuccess = () => {
        images.value = []
        resolve()
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Create a new conversation
   */
  const createConversation = async (model?: string): Promise<string> => {
    const conversation: Conversation = {
      id: crypto.randomUUID(),
      title: '新对话',
      model: model || 'gpt-4o',
      messages: [],
      createdAt: Date.now(),
      updatedAt: Date.now(),
      totalCost: 0,
      totalTokens: 0
    }

    await saveConversation(conversation)
    currentConversationId.value = conversation.id
    return conversation.id
  }

  /**
   * Update conversation title
   */
  const updateConversationTitle = async (conversationId: string, title: string) => {
    const conv = conversations.value.find((c) => c.id === conversationId)
    if (conv) {
      conv.title = title
      conv.updatedAt = Date.now()
      await saveConversation(conv)
    }
  }

  /**
   * Get total cost across all conversations
   */
  const getTotalCost = () => {
    return conversations.value.reduce((sum, conv) => sum + conv.totalCost, 0)
  }

  /**
   * Get default model from localStorage
   */
  const getDefaultModel = async (): Promise<string | null> => {
    return localStorage.getItem('sub2api_chat_default_model')
  }

  /**
   * Set default model to localStorage
   */
  const setDefaultModel = (model: string) => {
    localStorage.setItem('sub2api_chat_default_model', model)
  }

  return {
    conversations,
    currentConversationId,
    currentConversation,
    images,
    getTotalCost,
    loadConversations,
    saveConversation,
    deleteConversation,
    loadMessages,
    saveMessage,
    loadImages,
    saveImage,
    deleteImage,
    clearAllImages,
    createConversation,
    updateConversationTitle,
    getDefaultModel,
    setDefaultModel
  }
}
