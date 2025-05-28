<template>
  <Card class="w-full max-w-md">
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-lg">{{ submodule.name }}</CardTitle>
        <div class="flex items-center gap-2">
          <Badge :variant="getBranchVariant(submodule.currentBranch)" class="text-xs">
            {{ submodule.currentBranch }}
          </Badge>
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
      <Button @click="handleRcTag" variant="outline" size="sm" class="flex-1" :disabled="loading">
        <span class="mr-1">🔄</span>
        Tag RC
      </Button>
      <Button @click="handleProdTag" variant="default" size="sm" class="flex-1" :disabled="loading">
        <span class="mr-1">🚀</span>
        Tag Prod
      </Button>
    </CardFooter>
  </Card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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

const emit = defineEmits<{
  rcTag: [submodule: Submodule]
  prodTag: [submodule: Submodule]
}>()

const getBranchVariant = (branch: string) => {
  if (branch === 'main' || branch === 'master') return 'default'
  if (branch.startsWith('develop')) return 'secondary'
  if (branch.startsWith('feature/')) return 'outline'
  if (branch.startsWith('hotfix/')) return 'destructive'
  return 'outline'
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

const handleRcTag = () => {
  emit('rcTag', props.submodule)
}

const handleProdTag = () => {
  emit('prodTag', props.submodule)
}
</script>