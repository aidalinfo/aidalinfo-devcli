<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import SubmoduleCard from '@/components/SubmoduleCard.vue'
import TagDialog from '@/components/TagDialog.vue'
import { useSubmodules, type Submodule } from '@/composables/useSubmodules'

const { scanPath, submodules, loading, error, scanSubmodules, changeBranch, loadBranches, getDiff } = useSubmodules()

// TagDialog state
const tagDialogOpen = ref(false)
const selectedSubmodule = ref<Submodule | null>(null)
const tagType = ref<'rc' | 'prod'>('rc')
const submoduleRefs = ref<any[]>([])

const handleMergeToCycle = (submodule: Submodule) => {
  console.log('Merge to cycle for:', submodule.name)
  // TODO: Implement merge to cycle logic
}

const handleViewDiff = async (submodule: Submodule) => {
  const diff = await getDiff(submodule)
  const index = submodules.value.findIndex(s => s.path === submodule.path)
  if (index !== -1 && submoduleRefs.value[index]) {
    submoduleRefs.value[index].openDiffDialog(diff)
  }
}

const handleTagCreated = () => {
  // Refresh submodules to update tag information
  scanSubmodules()
}
</script>

<template>
  <div class="min-h-screen w-full p-4">
    <div class="container mx-auto max-w-7xl">
      <h1 class="text-3xl font-bold mb-8">Merger Management</h1>

      <!-- Scan Form -->
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
            v-for="(submodule, index) in submodules"
            :key="submodule.path"
            :ref="(el) => { if (el) submoduleRefs[index] = el }"
            :submodule="submodule"
            @change-branch="changeBranch"
            @load-branches="loadBranches"
            @view-diff="handleViewDiff"
            @merge-to-cycle="handleMergeToCycle"
          />
        </div>
      </div>

      <div v-else-if="!loading && scanPath" class="flex flex-1 items-center justify-center h-64">
        <p class="text-gray-600 text-xl">Aucun submodule trouvé dans ce projet</p>
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
