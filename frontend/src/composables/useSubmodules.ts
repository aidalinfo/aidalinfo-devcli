import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { ListSubmodules, CleanSubmodules, GetCurrentBranch, GitStatus, GetLastCommits, GetLastTags, GetBranches, ChangeBranch, GetPendingChanges } from '@/../wailsjs/go/main/App'
import type { backend } from '@/../wailsjs/go/models'

// Use the backend Commit type
type Commit = backend.Commit

export interface Tags {
  vTags: string[]
  rcTags: string[]
}

export interface Submodule {
  name: string
  path: string
  currentBranch: string
  branches?: string[]
  gitStatus?: string
  lastCommits?: Commit[]
  tags?: Tags
  pendingChanges?: string
}

export function useSubmodules() {
  const scanPath = ref('')
  const submodules = ref<Submodule[]>([])
  const loading = ref(false)
  const error = ref('')

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
      toast.error('Veuillez entrer un chemin à scanner')
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
          const pendingChanges = await GetPendingChanges(path)
          // Branches will be loaded on demand
          
          detailedSubmodules.push({
            name,
            path,
            currentBranch: currentBranch || 'unknown',
            branches: [], // Initialize empty
            gitStatus: gitStatus || undefined,
            pendingChanges: pendingChanges || undefined
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
      
      // ... existing code ...
      
      submodules.value = detailedSubmodules
      toast.success(`Scan terminé: ${detailedSubmodules.length} submodules trouvés`)
      
    } catch (err: any) {
      console.error('Error scanning submodules:', err)
      error.value = err?.message || String(err)
      toast.error(`Erreur lors du scan: ${error.value}`)
    } finally {
      loading.value = false
    }
  }

  const loadBranches = async (submodule: Submodule) => {
    try {
      const branches = await GetBranches(submodule.path)
      
      const index = submodules.value.findIndex(s => s.path === submodule.path)
      if (index !== -1) {
        submodules.value[index] = {
          ...submodules.value[index],
          branches: branches || []
        }
      }
    } catch (err) {
      console.error(`Error loading branches for ${submodule.name}:`, err)
      toast.error(`Erreur lors du chargement des branches: ${err}`)
    }
  }

  const changeBranch = async (submodule: Submodule, newBranch: string) => {
    try {
      await ChangeBranch(submodule.path, newBranch)
      toast.success(`Branche changée pour ${submodule.name} : ${newBranch}`)
      // Refresh info for this submodule
      const currentBranch = await GetCurrentBranch(submodule.path)
      const gitStatus = await GitStatus(submodule.path)
      // Refresh branches list as well to ensure it's up to date
      const branches = await GetBranches(submodule.path)
      
      // Update local state
      const index = submodules.value.findIndex(s => s.path === submodule.path)
      if (index !== -1) {
        submodules.value[index] = {
          ...submodules.value[index],
          currentBranch,
          gitStatus,
          branches
        }
      }
    } catch (err: any) {
      console.error(`Error changing branch for ${submodule.name}:`, err)
      toast.error(`Erreur lors du changement de branche : ${err}`)
    }
  }

  const getDiff = async (submodule: Submodule): Promise<string> => {
    try {
      // @ts-ignore
      const diff = await window['go']['main']['App']['GetDiff'](submodule.path)
      return diff
    } catch (err: any) {
      console.error(`Error getting diff for ${submodule.name}:`, err)
      toast.error(`Erreur lors de la récupération du diff : ${err}`)
      return ''
    }
  }

  return {
    scanPath,
    submodules,
    loading,
    error,
    scanSubmodules,
    savePath,
    changeBranch,
    loadBranches,
    getDiff
  }
}
