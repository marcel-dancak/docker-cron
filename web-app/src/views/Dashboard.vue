<template>
  <v-layout class="dashboard column py-3">
    <v-card class="tasks-card mx-2">
      <v-toolbar
        class="shrink"
        color="blue-grey darken-3"
        dark dense
      >
        <v-toolbar-title>Tasks</v-toolbar-title>
      </v-toolbar>

      <v-layout class="list column scrollable py-2">
        <task-info
          v-for="task in tasksList"
          :key="task.name"
          :task=task
          class="shrink mx-2 my-1"
        />
      </v-layout>
    </v-card>
  </v-layout>
</template>

<script>
import TaskInfo from '@/components/TaskInfo.vue'

export default {
  name: 'Tasks',
  components: { TaskInfo },
  computed: {
    tasksList () {
      return Object.values(this.$root.tasks)
    }
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  grid-column: 2 / 3;
  grid-row: 1 / 2;
  width: 100%;
  display: grid;
  grid-template-columns: 100%;
  grid-template-rows: 100%;
}

.tasks-card {
  opacity: 0.96;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
  max-height: 100%;
  align-self: start;

  @media (min-width: 960px) {
    width: 600px;
  }

  .v-divider {
    grid-column: 1 / 4;
  }
  .grid {
    display: grid;
    grid-template-columns: 1fr auto auto;
    > div {
      display: contents;
    }
  }
  .list {
    > *:not(:last-child) {
      border-bottom: 1px solid #ccc;
    }
  }
}
</style>
