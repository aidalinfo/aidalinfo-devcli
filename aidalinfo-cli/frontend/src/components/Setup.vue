<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ListSubmodules, CleanSubmodules, InstallSubmodules, InstallNpmDependencies } from '../../wailsjs/go/main/App'
import { toast } from 'vue-sonner'

const scanPath = ref('')
const submodules = ref<{ name: string; path: string }[]>([])
const branchesInput = ref('')
const branches = ref<string[]>([])
const loading = ref(false)
const error = ref('')
const npmInstall = ref(false)

const scanSubmodules = async () => {
  if (!scanPath.value.trim()) {
    toast.error('Veuillez entrer un chemin à scanner')
    return
  }
  loading.value = true
  error.value = ''
  submodules.value = []
  try {
    const submodulePaths = await ListSubmodules(scanPath.value)
    const submoduleNames = await CleanSubmodules(submodulePaths)
    submodules.value = submodulePaths.map((path, i) => ({ name: submoduleNames[i], path }))
    toast.success('Scan des submodules terminé')
  } catch (err) {
    toast.error(`Erreur lors du scan: ${err}`)
  } finally {
    loading.value = false
  }
}

const addBranch = () => {
  const branch = branchesInput.value.trim()
  if (branch && !branches.value.includes(branch)) {
    branches.value.push(branch)
    branchesInput.value = ''
  }
}

const removeBranch = (branch: string) => {
  branches.value = branches.value.filter(b => b !== branch)
}

const handleBranchesInputKeyup = (e: KeyboardEvent) => {
  if (e.key === 'Enter') {
    addBranch()
  }
}

const handleSetup = async () => {
  if (branches.value.length === 0) {
    toast.error('Ajoutez au moins une branche avant de lancer le setup.')
    return
  }
  loading.value = true
  try {
    await InstallSubmodules(scanPath.value || '.', branches.value)
    toast.success('Checkout des submodules terminé !')
    if (npmInstall.value) {
      await InstallNpmDependencies(scanPath.value || '.', true)      
      toast.success('npm install terminé sur tous les modules !')

    }
  } catch (err: any) {
    const msg = err?.message || err?.toString() || 'Erreur lors du setup'
    toast.error(msg, { duration: 10000 })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Setup des submodules</CardTitle>
      <CardDescription>
        Entrez le chemin du projet, scannez les submodules, ajoutez une ou plusieurs branches à essayer (appuyez sur Entrée ou cliquez sur Ajouter), puis lancez le checkout et npm install si besoin.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div class="flex gap-4 items-end mb-4">
        <div class="flex-1">
          <Label for="scan-path">Chemin du projet</Label>
          <Input id="scan-path" v-model="scanPath" placeholder="/chemin/vers/projet" @keyup.enter="scanSubmodules" />
        </div>
        <Button @click="scanSubmodules" :disabled="loading || !scanPath.trim()">
          {{ loading ? 'Scan en cours...' : 'Scanner' }}
        </Button>
      </div>
      <div v-if="submodules.length > 0">
        <div class="mb-4">
          <Label for="branches-input">Branches à essayer</Label>
          <div class="flex gap-2">
            <Input id="branches-input" v-model="branchesInput" placeholder="main ou feature/ma-branche" @keyup="handleBranchesInputKeyup" />
            <Button type="button" @click="addBranch" :disabled="!branchesInput.trim()">Ajouter</Button>
          </div>
          <div v-if="branches.length" class="flex flex-wrap gap-2 mt-2">
            <span v-for="branch in branches" :key="branch" class="inline-flex items-center bg-gray-100 rounded px-2 py-1 text-sm">
              {{ branch }}
              <button @click="removeBranch(branch)" class="ml-1 text-red-500 hover:text-red-700" title="Retirer">&times;</button>
            </span>
          </div>
        </div>
        <div class="mb-4">
          <Label>Submodules trouvés :</Label>
          <ul class="list-disc ml-6 text-sm">
            <li v-for="sub in submodules" :key="sub.path">
              <span class="font-semibold">{{ sub.name }}</span>
              <span class="text-xs text-gray-500 ml-2">({{ sub.path }})</span>
            </li>
          </ul>
        </div>
        <div class="flex items-center gap-2 mt-6">
          <input type="checkbox" id="npm-install" v-model="npmInstall" class="accent-blue-600" />
          <Label for="npm-install">Lancer npm install sur tous les modules après checkout</Label>
        </div>
        <Button class="mt-6 w-full" @click="handleSetup" :disabled="loading || branches.length === 0">
          {{ loading ? 'Traitement...' : 'Lancer le setup' }}
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
