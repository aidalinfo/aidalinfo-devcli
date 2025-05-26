<template>
  <SidebarProvider>
    <div class="flex min-h-screen w-full">
      <Navigation />
      <div class="flex-1 flex flex-col">
        <div class="flex items-center gap-4 p-4">
          <SidebarTrigger />
          <img id="logo" alt="Aidalinfo CLI" src="@/assets/aidalinfo.svg" class="h-7 w-auto"/>
        </div>
        <router-view />
      </div>
    </div>
  </SidebarProvider>
</template>

<script setup lang="ts">
import Navigation from '@/components/Navigation.vue'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
</script>

<style>
/* Styles globaux éventuels */
</style>
