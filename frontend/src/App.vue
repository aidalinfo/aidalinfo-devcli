<template>
  <SidebarProvider>
    <div class="flex min-h-screen w-full">
      <Navigation />
      <div class="flex-1 flex flex-col">
        <!-- Barre supérieure fixe avec espacement pour futurs composants -->
        <div class="sticky top-0 z-10 flex items-center justify-between gap-4 p-4 bg-background border-b">
          <div class="flex items-center gap-4">
            <SidebarTrigger />
            <img id="logo" alt="Aidalinfo CLI" src="@/assets/aidalinfo.svg" class="h-7 w-auto"/>
          </div>
          <!-- Espace réservé pour futurs composants à droite -->
          <div class="flex items-center gap-2">
            <Action />
            <!-- Placeholder pour futurs composants -->
          </div>
        </div>
        <!-- Contenu principal avec scroll -->
         
        <div class="flex-1 overflow-auto">
          <router-view />
        </div>
      </div>
    </div>
    <Toaster />
  </SidebarProvider>
</template>

<script setup lang="ts">
import Navigation from '@/components/Navigation.vue'
import Action from '@/components/Action.vue'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Toaster } from '@/components/ui/sonner'
import 'vue-sonner/style.css'
</script>

<style>
/* Styles globaux éventuels */
</style>
