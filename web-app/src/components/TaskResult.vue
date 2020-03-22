<template>
  <v-layout :class="{ small }">
    <v-icon
      :color="color"
      v-text="icon"
    />
    <span
      v-if="!hideText"
      v-text="resultText"
      class="ml-1"
    />
  </v-layout>
</template>

<script>
const StatusIcons = {
  success: 'check_circle',
  error: 'error',
  crashed: 'notification_important',
  pending: ' '
}
const StatusColors = {
  success: 'green',
  error: 'deep-orange',
  crashed: 'red darken-2',
  pending: ''
}
const StatusText = {
  success: 'Success',
  error: 'Error',
  crashed: 'Failure',
  pending: ''
}

export default {
  props: {
    stats: Object,
    small: Boolean,
    hideText: Boolean
  },
  computed: {
    status () {
      if (!this.stats) {
        return 'pending'
      }
      return this.stats.crashed ? 'crashed' : this.stats.status === 0 ? 'success' : 'error'
    },
    icon () {
      return StatusIcons[this.status]
    },
    color () {
      return StatusColors[this.status]
    },
    resultText () {
      return StatusText[this.status]
    }
  }
}
</script>

<style lang="scss" scoped>
.small {
  .v-icon {
    font-size: 19px;
  }
  span {
    font-size: 80%;
  }
}
</style>
