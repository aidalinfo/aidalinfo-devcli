// Notification utility - migrated to vue-sonner toast system
// This file is deprecated - use toast from 'vue-sonner' directly

import { toast } from 'vue-sonner'

export function showNotification(message: string, type: 'success' | 'error' | 'info' = 'info') {
  // Migrated to vue-sonner toast system
  switch (type) {
    case 'success':
      toast.success(message)
      break
    case 'error':
      toast.error(message)
      break
    case 'info':
    default:
      toast.info(message)
      break
  }
}