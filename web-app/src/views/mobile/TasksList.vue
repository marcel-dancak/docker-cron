<template>
  <v-layout class="column page px-1 py-1">
    <task-info
      v-for="t in tasksList"
      :key="t.name"
      :task="t"
    />
  </v-layout>
</template>

<script>
import TaskInfo from '@/components/mobile/TaskInfo.vue'

export default {
  components: { TaskInfo },
  computed: {
    tasks () {
      return this.$root.tasks
    },
    tasksList () {
      return Object.values(this.tasks)
    }
  }
}
</script>

<style lang="scss" scoped>
.page {
  // background-color: #555;
  max-width: 600px;
  min-height: 0;
  max-height: 100%;
  > * {
    margin: 8px 4px;
  }
  width: 600px;
  @media (max-width: 600px) {
    align-self: center;
    width: 100%;
  }
}
</style>
