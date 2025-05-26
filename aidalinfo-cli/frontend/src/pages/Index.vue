<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import SubmoduleCard from '@/components/SubmoduleCard.vue'
import TagDialog from '@/components/TagDialog.vue'
import { ListSubmodules, CleanSubmodules, GetCurrentBranch, GitStatus, GetLastCommits, GetLastTags } from '@/../wailsjs/go/main/App'
import type { backend } from '@/../wailsjs/go/models'
import { SidebarTrigger } from '@/components/ui/sidebar'

// Use the backend Commit type
// ...existing code...
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

const scanPath = ref('')
const submodules = ref<Submodule[]>([])
const loading = ref(false)
const error = ref('')

// TagDialog state
const tagDialogOpen = ref(false)
const selectedSubmodule = ref<Submodule | null>(null)
const tagType = ref<'rc' | 'prod'>('rc')

// Load cached path on mount
onMounted(() => {
  const cachedPath = localStorage.getItem('aidalinfo-scan-path')
  if (cachedPath) {
    scanPath.value = cachedPath
  }
})

const savePath = () => {
  if (scanPath.value) {
    localStorage.setItem('aidalinfo-scan-path', scanPath.value)
  }
}

const scanSubmodules = async () => {
  if (!scanPath.value.trim()) {
    error.value = 'Veuillez entrer un chemin à scanner'
    return
  }

  loading.value = true
  error.value = ''
  
  try {
    // Save path to cache
    savePath()
    
    // Get submodule paths
    const submodulePaths = await ListSubmodules(scanPath.value)
    const submoduleNames = await CleanSubmodules(submodulePaths)
    
    // Get detailed info for each submodule
    const detailedSubmodules: Submodule[] = []
    
    for (let i = 0; i < submodulePaths.length; i++) {
      const path = submodulePaths[i]
      const name = submoduleNames[i]
      
      try {
        const currentBranch = await GetCurrentBranch(path)
        const gitStatus = await GitStatus(path)
        
        detailedSubmodules.push({
          name,
          path,
          currentBranch: currentBranch || 'unknown',
          gitStatus: gitStatus || undefined
        })
      } catch (err) {
        console.error(`Error getting info for ${name}:`, err)
        detailedSubmodules.push({
          name,
          path,
          currentBranch: 'error',
          gitStatus: `Error: ${err}`
        })
      }
    }
    
    // Get last commits for all submodules
    try {
      const commits = await GetLastCommits(submodulePaths)
      
      // Group commits by submodule
      const commitsBySubmodule = commits.reduce((acc, commit) => {
        if (!acc[commit.Submodule]) {
          acc[commit.Submodule] = []
        }
        acc[commit.Submodule].push(commit)
        return acc
      }, {} as Record<string, Commit[]>)
      
      // Add commits to submodules
      detailedSubmodules.forEach(submodule => {
        submodule.lastCommits = commitsBySubmodule[submodule.name] || []
      })
    } catch (err) {
      console.error('Error getting commits:', err)
    }
    
    // Get tags for each submodule
    for (const submodule of detailedSubmodules) {
      try {
        const tagsResult = await GetLastTags(submodule.path)
        submodule.tags = {
          vTags: Array.isArray(tagsResult.vTags) ? tagsResult.vTags : [],
          rcTags: Array.isArray(tagsResult.rcTags) ? tagsResult.rcTags : []
        }
      } catch (err) {
        console.error(`Error getting tags for ${submodule.name}:`, err)
        submodule.tags = {
          vTags: [],
          rcTags: []
        }
      }
    }
    
    submodules.value = detailedSubmodules
    
  } catch (err) {
    console.error('Error scanning submodules:', err)
    error.value = `Erreur lors du scan: ${err}`
  } finally {
    loading.value = false
  }
}

const handleRcTag = (submodule: Submodule) => {
  selectedSubmodule.value = submodule
  tagType.value = 'rc'
  tagDialogOpen.value = true
}

const handleProdTag = (submodule: Submodule) => {
  selectedSubmodule.value = submodule
  tagType.value = 'prod'
  tagDialogOpen.value = true
}

const handleTagCreated = () => {
  // Refresh submodules to update tag information
  scanSubmodules()
}
</script>

<template>
  <div class="min-h-screen w-full">
    <!-- Si aucun chemin n'est scanné, afficher le formulaire centré en plein écran -->
    <div v-if="!scanPath" class="flex flex-col justify-center items-center min-h-screen w-full bg-white">
      <Card>
        <CardHeader>
          <CardTitle>Scanner les submodules</CardTitle>
          <CardDescription>
            Entrez le chemin du projet à scanner pour découvrir les submodules
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="flex gap-4 items-end">
            <div class="flex-1">
              <Label for="scan-path">Chemin du projet</Label>
              <Input 
                id="scan-path"
                v-model="scanPath" 
                placeholder="C:\\Users\\...\\mon-projet"
                @keyup.enter="scanSubmodules"
              />
            </div>
            <Button 
              @click="scanSubmodules" 
              :disabled="loading || !scanPath.trim()"
              class="mb-0"
            >
              {{ loading ? 'Scan en cours...' : 'Scanner' }}
            </Button>
          </div>
          <div v-if="error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded text-red-700">
            {{ error }}
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Si un chemin est scanné, afficher la vue principale avec sidebar, header, résultats -->
    <div v-else class="w-full p-4">
      <div class="container mx-auto max-w-7xl">
        <!-- Scan Form (toujours visible pour changer de chemin) -->
        <Card class="mb-8">
          <CardHeader>
            <CardTitle>Scanner les submodules</CardTitle>
            <CardDescription>
              Entrez le chemin du projet à scanner pour découvrir les submodules
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex gap-4 items-end">
              <div class="flex-1">
                <Label for="scan-path">Chemin du projet</Label>
                <Input 
                  id="scan-path"
                  v-model="scanPath" 
                  placeholder="C:\\Users\\...\\mon-projet"
                  @keyup.enter="scanSubmodules"
                />
              </div>
              <Button 
                @click="scanSubmodules" 
                :disabled="loading || !scanPath.trim()"
                class="mb-0"
              >
                {{ loading ? 'Scan en cours...' : 'Scanner' }}
              </Button>
            </div>
            <div v-if="error" class="mt-4 p-3 bg-red-50 border border-red-200 rounded text-red-700">
              {{ error }}
            </div>
          </CardContent>
        </Card>
        <!-- Results -->
        <div v-if="loading" class="text-center py-12">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-4 text-gray-600">Scan des submodules en cours...</p>
        </div>
        <div v-else-if="submodules.length > 0" class="space-y-6">
          <div class="flex items-center justify-between">
            <h2 class="text-2xl font-bold text-gray-800">
              Submodules trouvés ({{ submodules.length }})
            </h2>
            <Button variant="outline" @click="scanSubmodules" :disabled="loading">
              Actualiser
            </Button>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <SubmoduleCard
              v-for="submodule in submodules"
              :key="submodule.path"
              :submodule="submodule"
              @rc-tag="handleRcTag"
              @prod-tag="handleProdTag"
            />
          </div>
        </div>
        <div v-else-if="!loading && scanPath" class="flex flex-1 items-center justify-center h-full">
          <p class="text-gray-600 text-xl">Aucun submodule trouvé dans ce projet</p>
        </div>
      </div>
      <!-- Tag Dialog -->
      <TagDialog
        v-if="selectedSubmodule"
        :open="tagDialogOpen"
        @update:open="tagDialogOpen = $event"
        :submodule-path="selectedSubmodule.path"
        :submodule-name="selectedSubmodule.name"
        :tag-type="tagType"
        @tag-created="handleTagCreated"
      />
    </div>
  </div>
</template>
