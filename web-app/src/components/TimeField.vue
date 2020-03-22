<template functional>
  <v-layout
    class="align-center shrink"
    :class="[data.class, data.staticClass]"
  >
    <v-icon size="19">access_time</v-icon>
    <small class="ml-1">{{ props.value | time }}</small>
  </v-layout>
</template>

<script>
export default {
  props: {
    value: String
  }
}
</script>
