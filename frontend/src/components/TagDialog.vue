<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[600px] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle>{{ tagType === 'rc' ? 'RC Tag' : 'Production Tag' }}</DialogTitle>
        <DialogDescription>
          Créer un nouveau tag {{ tagType === 'rc' ? 'RC (Release Candidate)' : 'de production' }} pour le sous-module {{ submoduleName }}
        </DialogDescription>
      </DialogHeader>
      
      <div class="grid gap-4 py-4 flex-1 overflow-hidden">
        <!-- Version suggérée -->
        <div class="grid grid-cols-4 items-center gap-4">
          <Label for="version" class="text-right">Version</Label>
          <Input
            id="version"
            v-model="version"
            class="col-span-3"
            :placeholder="suggestedVersion"
          />
        </div>
        
        <!-- Message du tag -->
        <div class="grid grid-cols-4 items-center gap-4">
          <Label for="message" class="text-right">Message</Label>
          <Input
            id="message"
            v-model="message"
            class="col-span-3"
            placeholder="Description du tag"
          />
        </div>
        
        <!-- Historique des tags -->
        <div class="grid grid-cols-4 gap-4 flex-1 overflow-hidden">
          <Label class="text-right">Historique</Label>
          <div class="col-span-3 border rounded-md overflow-hidden flex flex-col h-60">
            <div class="bg-muted px-3 py-2 text-sm font-medium">
              Tags {{ tagType === 'rc' ? 'RC' : 'Production' }} récents
            </div>
            <div class="flex-1 overflow-y-auto p-2">
              <div v-if="loading" class="text-center py-4 text-muted-foreground">
                Chargement des tags...
              </div>
              <div v-else-if="relevantTags.length === 0" class="text-center py-4 text-muted-foreground">
                Aucun tag {{ tagType === 'rc' ? 'RC' : 'de production' }} trouvé
              </div>
              <div v-else class="space-y-1">
                <div
                  v-for="tag in relevantTags"
                  :key="tag"
                  class="px-2 py-1 text-sm rounded text-secondary-foreground"
                >
                  {{ tag }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <DialogFooter>
        <Button variant="outline" @click="$emit('update:open', false)">
          Annuler
        </Button>
        <Button @click="createTag" :disabled="!version || !message || creating">
          {{ creating ? 'Création...' : 'Créer et Push' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { GetLastTags, CreateTag } from '../../wailsjs/go/main/App'
import { toast } from 'vue-sonner'
interface Props {
  open: boolean
  submodulePath: string
  submoduleName: string
  tagType: 'rc' | 'prod'
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'tag-created': []
}>()

const version = ref('')
const message = ref('')
const loading = ref(false)
const creating = ref(false)
const vTags = ref<string[]>([])
const rcTags = ref<string[]>([])

const relevantTags = computed(() => {
  return props.tagType === 'rc' ? rcTags.value : vTags.value
})

const suggestedVersion = computed(() => {
  if (props.tagType === 'rc') {
    // Pour RC, suggérer rc-X.X.X
    const lastRcTag = rcTags.value[0]
    if (lastRcTag) {
      // Extraire le numéro de version et l'incrémenter
      const match = lastRcTag.match(/rc-v(\d+)\.(\d+)\.(\d+)\.(\d+)/)
      if (match) {
        const [, major, minor, patch, build] = match
        return `rc-v${major}.${minor}.${patch}.${parseInt(build) + 1}`
      }
    }
    return 'rc-1.0.0'
  } else {
    // Pour prod, suggérer vX.X.XXXX (padding conservé)
    const lastVTag = vTags.value[0]
    if (lastVTag) {
      // Match v1.1.0039 ou v1.1.39 etc.
      const match = lastVTag.match(/v(\d+)\.(\d+)\.(\d+)/)
      if (match) {
        const [, major, minor, patch] = match
        // Conserver le padding du patch
        const patchLen = patch.length
        const nextPatch = (parseInt(patch, 10) + 1).toString().padStart(patchLen, '0')
        return `v${major}.${minor}.${nextPatch}`
      }
    }
    return 'v1.0.0'
  }
})

const loadTags = async () => {
  if (!props.submodulePath) return
  
  loading.value = true
  try {
    const tagsResult = await GetLastTags(props.submodulePath)
    vTags.value = Array.isArray(tagsResult.vTags) ? tagsResult.vTags : []
    rcTags.value = Array.isArray(tagsResult.rcTags) ? tagsResult.rcTags : []
  } catch (error) {
    console.error('Erreur lors du chargement des tags:', error)
    toast.error("Impossible de charger l'historique des tags")
  } finally {
    loading.value = false
  }
}

const createTag = async () => {
  if (!version.value || !message.value) return
  
  creating.value = true
  try {
    await CreateTag(props.submodulePath, version.value, message.value)
    toast({
      title: "Succès",
      description: `Tag ${version.value} créé et pushé avec succès`,
    })
    emit('tag-created')
    emit('update:open', false)
    
    // Reset form
    version.value = ''
    message.value = ''
  } catch (error) {
    console.error('Erreur lors de la création du tag:', error)
    toast({
      title: "Erreur",
      description: `Échec de la création du tag: ${error}`,
      variant: "destructive",
    })
  } finally {
    creating.value = false
  }
}

// Charger les tags quand le dialog s'ouvre
watch(() => props.open, (newOpen) => {
  if (newOpen) {
    loadTags().then(() => {
      // Mettre le dernier tag comme valeur par défaut dans l'input
      if (props.tagType === 'rc') {
        version.value = rcTags.value[0] || ''
      } else {
        version.value = vTags.value[0] || ''
      }
    })
  }
})

// Initialiser au montage si le dialog est déjà ouvert
onMounted(() => {
  if (props.open) {
    loadTags().then(() => {
      if (props.tagType === 'rc') {
        version.value = rcTags.value[0] || ''
      } else {
        version.value = vTags.value[0] || ''
      }
    })
  }
})
</script>
