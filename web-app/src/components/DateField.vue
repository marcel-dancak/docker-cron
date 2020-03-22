<template functional>
  <v-layout
    class="align-center shrink"
    :class="[data.class, data.staticClass]"
  >
    <v-icon size="19">date_range</v-icon>
    <small class="ml-1">{{ props.value | date }}</small>
  </v-layout>
</template>

<script>
export default {
  props: {
    value: String
  }
}
</script>
