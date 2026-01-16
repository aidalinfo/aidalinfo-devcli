<template>
  <Card class="w-full max-w-md">
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-lg">{{ submodule.name }}</CardTitle>
        <div class="flex items-center gap-2">
          <Select 
            :model-value="submodule.currentBranch"
            @update:model-value="handleBranchChange"
            @update:open="handleDropdownOpen">
            <SelectTrigger class="w-[180px] h-8 text-xs">
              <SelectValue :placeholder="submodule.currentBranch" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectLabel>Branches</SelectLabel>
                <SelectItem v-if="submodule.branches && submodule.branches.length > 0" 
                  v-for="branch in submodule.branches" :key="branch" :value="branch">
                  {{ branch }}
                </SelectItem>
                <SelectItem v-else value="loading" disabled>
                  Chargement...
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>

          <div v-if="submodule.lastCommits && submodule.lastCommits.length > 0"
            class="cursor-pointer p-1 rounded-full hover:bg-gray-100 transition-colors" :title="getLastCommitTooltip()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              stroke-linecap="round" stroke-linejoin="round" class="text-gray-500">
              <circle cx="12" cy="12" r="10" />
              <path d="m9 12 2 2 4-4" />
            </svg>
          </div>
        </div>
      </div>
      <CardDescription>
        {{ submodule.path }}
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      <!-- Tags -->
      <div class="space-y-2" v-if="submodule.tags">
        <h4 class="text-sm font-medium">Tags</h4>
        <div class="flex flex-wrap gap-1">
          <Badge v-for="tag in submodule.tags.vTags?.slice(0, 3)" :key="tag" variant="secondary" class="text-xs">
            {{ tag }}
          </Badge>
          <Badge v-for="tag in submodule.tags.rcTags?.slice(0, 2)" :key="tag" variant="outline" class="text-xs">
            {{ tag }}
          </Badge>
        </div>
      </div>

      <!-- Pending Changes -->
      <div v-if="submodule.pendingChanges" class="space-y-2">
        <h4 class="text-sm font-medium text-orange-600">Modifications en attente</h4>
        <div class="text-xs text-orange-700 bg-orange-50 p-2 rounded">
          <pre class="whitespace-pre-wrap">{{ submodule.pendingChanges }}</pre>
        </div>
      </div>
    </CardContent>

    <CardFooter class="flex gap-2">
      <Button @click="handleViewDiff" variant="outline" size="sm" class="flex-1" :disabled="loading || !submodule.pendingChanges">
        <span class="mr-1">📄</span>
        Changements <span v-if="getPendingChangesCount() > 0">({{ getPendingChangesCount() }})</span>
      </Button>
      <Button @click="handleMergeToCycle" variant="default" size="sm" class="flex-1" :disabled="loading">
        <span class="mr-1">🔀</span>
        Merge to cycle
      </Button>
      <Dialog v-model:open="diffDialogOpen">
        <DialogContent class="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>Modifications : {{ submodule.name }}</DialogTitle>
            <DialogDescription>
              Diff des fichiers modifiés
            </DialogDescription>
          </DialogHeader>
          <div class="flex-1 overflow-auto bg-slate-950 p-4 rounded-md mt-2">
            <pre class="text-xs text-green-400 font-mono whitespace-pre-wrap">{{ diffContent }}</pre>
          </div>
        </DialogContent>
      </Dialog>
    </CardFooter>
  </Card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectGroup, SelectLabel } from '@/components/ui/select'

import type { backend } from '../../wailsjs/go/models'

// Use the backend Commit type
type Commit = backend.Commit

interface Tags {
  vTags: string[]
  rcTags: string[]
}

interface Submodule {
  name: string
  path: string
  currentBranch: string
  branches?: string[]
  gitStatus?: string
  lastCommits?: Commit[]
  tags?: Tags
  pendingChanges?: string
}

interface Props {
  submodule: Submodule
}

const props = defineProps<Props>()
const loading = ref(false)
const diffDialogOpen = ref(false)
const diffContent = ref('')

const emit = defineEmits<{
  changeBranch: [submodule: Submodule, branch: string]
  loadBranches: [submodule: Submodule]
  viewDiff: [submodule: Submodule]
  mergeToCycle: [submodule: Submodule]
}>()

const getBranchVariant = (branch: string) => {
  if (branch === 'main' || branch === 'master') return 'default'
  if (branch.startsWith('develop')) return 'secondary'
  if (branch.startsWith('feature/')) return 'outline'
  if (branch.startsWith('hotfix/')) return 'destructive'
  return 'outline'
}

// ... existing code ...

const handleMergeToCycle = () => {
  emit('mergeToCycle', props.submodule)
}

const getPendingChangesCount = () => {
  if (!props.submodule.pendingChanges) return 0
  return props.submodule.pendingChanges.trim().split('\n').length
}

const formatDate = (dateStr: string) => {
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString('fr-FR', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return dateStr
  }
}

const getLastCommitTooltip = () => {
  if (!props.submodule.lastCommits || props.submodule.lastCommits.length === 0) {
    return 'Aucun commit récent'
  }
  const lastCommit = props.submodule.lastCommits[0]
  return `Dernier commit:\n${lastCommit.Author}\n${formatDate(lastCommit.Date)}\n${lastCommit.Message}`
}



const handleBranchChange = (branch: any) => {
  emit('changeBranch', props.submodule, branch as string)
}

const handleDropdownOpen = (isOpen: boolean) => {
  if (isOpen && (!props.submodule.branches || props.submodule.branches.length === 0)) {
    emit('loadBranches', props.submodule)
  }
}

const handleViewDiff = () => {
  emit('viewDiff', props.submodule)
}

// Expose openDiffDialog to parent
defineExpose({
  openDiffDialog: (content: string) => {
    diffContent.value = content
    diffDialogOpen.value = true
  }
})
</script>