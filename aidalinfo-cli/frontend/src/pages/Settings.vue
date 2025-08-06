<template>
  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Serveurs S3/MinIO</CardTitle>
      <CardDescription>
        Gérez vos différents serveurs S3 et MinIO pour les backups et restaurations.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <S3ServerManager />
    </CardContent>
  </Card>

  <Card class="mx-5 mt-10">
    <CardHeader>
      <CardTitle>Serveurs MongoDB</CardTitle>
      <CardDescription>
        Gérez vos différents serveurs MongoDB pour la restauration et les dumps.
      </CardDescription>
    </CardHeader>
    <CardContent>
      <MongoServerManager />
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import MongoServerManager from '@/components/MongoServerManager.vue'
import S3ServerManager from '@/components/S3ServerManager.vue'
</script>