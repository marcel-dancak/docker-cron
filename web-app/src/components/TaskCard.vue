<template>
  <v-card>
    <v-toolbar
      color="blue-grey darken-3"
      elevation="1"
      height="72"
      dark
    >
      <v-list-item two-line class="px-1">
        <v-list-item-content>
          <v-list-item-title class="headline">
            <!-- <router-link :to="{ name: 'task', params: { name: task.name } }">{{ task.name }}</router-link> -->
            {{ task.name }}
          </v-list-item-title>
          <v-list-item-subtitle>{{ task.schedule }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
      <task-stats :task="task"/>
    </v-toolbar>
    <v-progress-linear v-if="running" indeterminate/>
    <v-layout class="my-2 px-3 py-1 align-center">
      <h5>Next:</h5>
      <date-field :value="task.next" class="ml-2"/>
      <time-field :value="task.next" class="ml-2"/>
      <v-spacer/>
      <v-btn
        rounded small
        :disabled="running"
        @click="runTask(task)"
      >
        Run
      </v-btn>
    </v-layout>
  </v-card>
</template>

<script>
import TaskStats from '@/components/TaskStats.vue'

export default {
  name: 'TaskCard',
  components: { TaskStats },
  props: {
    task: Object
  },
  computed: {
    stats () {
      return this.task.stats || []
    },
    running () {
      return this.stats.some(i => i.running)
    }
  },
  methods: {
    runTask (task) {
      this.$http.post(`/api/run/${task.name}`)
    }
  }
}
</script>

<style lang="scss" scoped>
::v-deep {
  a {
    text-decoration: none;
  }
}
</style>
