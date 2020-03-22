<template>
  <div class="card v-card">
    <v-layout
      class="header align-center px-3 grow blue-grey darken-3"
      @click="$router.push({ name: 'task', params: { name: task.name } })"
    >
      <v-list-item two-line class="px-0 py-0" dark>
        <v-list-item-content class="py-0">
          <v-list-item-title class="headline">
            {{ task.name }}
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ task.schedule }}
          </v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
      <task-stats :task="task"/>
    </v-layout>

    <v-layout class="last text--secondary column justify-start shrink px-3 my-2">
      <h5>Last</h5>
      <date-field :value="last && last.start_time"/>
      <time-field :value="last && last.start_time"/>
      <task-result :stats="last" small/>
    </v-layout>

    <v-layout class="next text--secondary column justify-start shrink px-3 my-2">
      <h5>Next</h5>
      <date-field :value="task.next"/>
      <time-field :value="task.next"/>
      <v-layout column py-1>
        <v-progress-linear v-if="running" indeterminate/>
      </v-layout>
    </v-layout>
  </div>
</template>

<script>
import findLast from 'lodash/findLast'
import TaskStats from '@/components/TaskStats.vue'
import TaskResult from '@/components/TaskResult.vue'

export default {
  name: 'TaskCard',
  components: { TaskStats, TaskResult },
  props: {
    task: Object
  },
  computed: {
    stats () {
      return this.task.stats || []
    },
    last () {
      return findLast(this.stats, r => !r.running)
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
.last, .next {
  .v-icon {
    margin-top: 2px;
    margin-bottom: 2px;
  }
}
.card {
  display: grid;
  grid-template-columns: 1fr auto;
  background-color: #eee;;

  .header {
    grid-column: 1 / -1;
    color: #fff;
  }
}
</style>
