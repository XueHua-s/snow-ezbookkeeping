<script setup lang="ts">
import type { Router } from 'framework7/types';
import AssistantConversation from '@/components/mobile/assistant/AssistantConversation.vue';
defineProps<{ f7router: Router.Router }>();
</script>

<template><AssistantConversation :f7router="f7router" /></template>
