<script setup lang="ts">
import WelcomeItem from './WelcomeItem.vue'
import DocumentationIcon from './icons/IconDocumentation.vue'
import ToolingIcon from './icons/IconTooling.vue'
import EcosystemIcon from './icons/IconEcosystem.vue'
import CommunityIcon from './icons/IconCommunity.vue'
import SupportIcon from './icons/IconSupport.vue'

const runGreet = () => {

}

</script>

<template>
  <WelcomeItem>
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>Welcome!</template>
    WAB is running now. To test it out click the 'Greet' button below:<br />
    <button @click="runGreet">Greet</button>
  </WelcomeItem>

  <!-- <WelcomeItem v-if="">
    <template #icon>
      <DocumentationIcon />
    </template>
    <template #heading>Welcome!</template>
    WAB is running now. To test it out click the 'Greet' button below:<br />
    <button>Greet</button>
  </WelcomeItem> -->
</template>
