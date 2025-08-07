<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>Mise à jour disponible</DialogTitle>
        <DialogDescription>
          Une nouvelle version est disponible pour Aidalinfo CLI.
        </DialogDescription>
      </DialogHeader>
      
      <div class="space-y-4 py-4">
        <div class="space-y-2">
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">Version actuelle:</span>
            <span class="text-sm font-medium">{{ currentVersion }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm text-muted-foreground">Nouvelle version:</span>
            <span class="text-sm font-medium text-green-600">{{ latestVersion }}</span>
          </div>
        </div>

        <div v-if="isDownloading" class="space-y-2">
          <div class="flex items-center space-x-2">
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
            <span class="text-sm">Téléchargement en cours...</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-2">
            <div class="bg-primary h-2 rounded-full transition-all duration-300" 
                 :style="`width: ${downloadProgress}%`"></div>
          </div>
        </div>

        <div v-if="updateError" class="text-sm text-red-600">
          {{ updateError }}
        </div>

        <div v-if="updateSuccess" class="text-sm text-green-600">
          Mise à jour réussie! L'application va redémarrer...
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false" :disabled="isDownloading">
          Annuler
        </Button>
        <Button @click="performUpdate" :disabled="isDownloading || updateSuccess">
          {{ isDownloading ? 'Téléchargement...' : 'Mettre à jour' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { CheckForUpdates, PerformUpdate } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { toast } from 'vue-sonner'

const props = defineProps<{
  modelValue: boolean
  currentVersion: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const isOpen = ref(props.modelValue)
const latestVersion = ref('')
const downloadURL = ref('')
const isDownloading = ref(false)
const downloadProgress = ref(0)
const updateError = ref('')
const updateSuccess = ref(false)

watch(() => props.modelValue, (newVal) => {
  isOpen.value = newVal
  if (newVal) {
    checkForUpdates()
  }
})

watch(isOpen, (newVal) => {
  emit('update:modelValue', newVal)
})

const checkForUpdates = async () => {
  try {
    const info = await CheckForUpdates()
    if (info) {
      latestVersion.value = info.latestVersion
      downloadURL.value = info.downloadUrl
    }
  } catch (error) {
    console.error('Erreur lors de la vérification des mises à jour:', error)
    updateError.value = 'Impossible de vérifier les mises à jour'
  }
}

const performUpdate = async () => {
  if (!downloadURL.value) return
  
  isDownloading.value = true
  downloadProgress.value = 0
  updateError.value = ''
  
  try {
    // Écouter les événements de progression (si implémentés)
    EventsOn('update:progress', (progress: number) => {
      downloadProgress.value = progress
    })
    
    await PerformUpdate(downloadURL.value)
    
    updateSuccess.value = true
    toast.success('Mise à jour réussie! L\'application va redémarrer...')
    
    // Attendre un peu avant de fermer
    setTimeout(() => {
      isOpen.value = false
      // L'application devrait redémarrer automatiquement
    }, 3000)
    
  } catch (error) {
    console.error('Erreur lors de la mise à jour:', error)
    updateError.value = `Échec de la mise à jour: ${error}`
  } finally {
    isDownloading.value = false
    EventsOff('update:progress')
  }
}
</script>