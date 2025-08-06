<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Tag, Settings, Database, Wrench, Download } from 'lucide-vue-next'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton
} from '@/components/ui/sidebar'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import UpdateDialog from './UpdateDialog.vue'
import { GetCurrentVersion, CheckForUpdates } from '../../wailsjs/go/main/App'
import { toast } from 'vue-sonner'

const items = [
  {
    title: 'Tag Management',
    route: '/',
    icon: Tag,
  },
  {
    title: 'Setup Management',
    route: '/setup',
    icon: Wrench,
  },
  {
    title: 'Backup Management',
    route: '/backup',
    icon: Database,
  },
  {
    title: 'Settings',
    route: '/settings',
    icon: Settings,
  },
]

const currentVersion = ref('1.0.0')
const showUpdateDialog = ref(false)
const updateAvailable = ref(false)

onMounted(async () => {
  try {
    currentVersion.value = await GetCurrentVersion()
    const updateInfo = await CheckForUpdates()
    if (updateInfo && updateInfo.updateAvailable) {
      updateAvailable.value = true
    }
  } catch (error) {
    console.error('Erreur lors de la récupération de la version:', error)
  }
})

const checkForUpdate = async () => {
  try {
    const updateInfo = await CheckForUpdates()
    if (updateInfo && updateInfo.updateAvailable) {
      updateAvailable.value = true
      showUpdateDialog.value = true
    } else {
      toast.info('Vous utilisez la dernière version disponible')
    }
  } catch (error) {
    toast.error('Impossible de vérifier les mises à jour')
  }
}
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Menu</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in items" :key="item.title">
              <SidebarMenuButton asChild>
                <router-link :to="item.route" class="flex items-center gap-2">
                  <component :is="item.icon" />
                  <span>{{ item.title }}</span>
                </router-link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
    
    <SidebarFooter>
      <Separator />
      <div class="p-3 space-y-2">
        <div class="flex items-center justify-between text-sm">
          <span class="text-muted-foreground">Version</span>
          <span class="font-medium">{{ currentVersion }}</span>
        </div>
        <Button 
          @click="checkForUpdate" 
          variant="outline" 
          size="sm"
          class="w-full"
          :class="{ 'animate-pulse': updateAvailable }"
        >
          <Download class="mr-2 h-4 w-4" />
          {{ updateAvailable ? 'Mise à jour disponible' : 'Vérifier les mises à jour' }}
        </Button>
      </div>
    </SidebarFooter>
  </Sidebar>
  
  <UpdateDialog 
    v-model="showUpdateDialog"
    :current-version="currentVersion"
  />
</template>
