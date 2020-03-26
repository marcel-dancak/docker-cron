<template>
  <v-layout
    v-if="task"
    class="task-page column py-1 grey lighten-2"
  >
    <task-card
      :task=task
      class="mx-2 my-2"
    />
    <v-card class="history mx-2 my-2">
      <v-toolbar
        dense
        color="blue-grey darken-3"
        elevation="1"
        dark
      >
        <v-icon class="mr-2">history</v-icon>
        <v-toolbar-title>History</v-toolbar-title>
      </v-toolbar>

      <v-list>
        <v-list-item v-if="!stats.length">
          <span class="text--secondary">No records</span>
        </v-list-item>
        <template v-for="run in stats">
          <v-list-item :key="run.id" class="px-3">
            <div class="id mr-1 grey darken-2 text-center">
              <small class="white--text font-weight-medium">{{ run.id }}</small>
            </div>
            <date-field :value="task.next" class="ml-2"/>
            <time-field :value="task.next" class="ml-2"/>

            <v-spacer/>
            <v-progress-circular
              v-if="run.running"
              size="20"
              width="2"
              color="primary"
              indeterminate
              class="mx-1"
            />
            <task-result v-else :stats="run" class="shrink" hide-text/>
            <!-- <v-icon
              v-else
              :color="run.status === 0 ? 'green' : 'red darken-1'"
              v-text="run.status === 0 ? 'check_circle' : 'error'"
              class="mx-1"
            /> -->
            <v-icon
              @click="toggleLogs(run.id)"
              v-text="'assignment'"
              class="mx-1"
              :disabled="!run.stderr_size && !run.stdout_size"
              :color="openLogs[run.id] ? 'primary' : ''"
            />
          </v-list-item>
          <v-expand-transition :key="`log-${run.id}`" class="py-2">
            <log-viewer
              v-if="openLogs[run.id] && fetchedLogs[run.id]"
              class="logs text--secondary px-3"
              :logs="fetchedLogs[run.id]"
            />
          </v-expand-transition>
        </template>
      </v-list>
    </v-card>
  </v-layout>
</template>

<script>
import TaskCard from '@/components/TaskCard.vue'
import LogViewer from '@/components/LogViewer.vue'
import TaskResult from '@/components/TaskResult.vue'

export default {
  name: 'Task',
  components: { TaskCard, TaskResult, LogViewer },
  props: {
    name: String
  },
  data () {
    return {
      logsId: null,
      fetchedLogs: {},
      openLogs: {}
    }
  },
  computed: {
    task () {
      return this.$root.tasks[this.name]
    },
    stats () {
      return this.task.stats.slice().reverse()
    }
  },
  methods: {
    async loadLog (id) {
      const resp = await this.$http.get(`/api/logs/${this.name}/${id}`)
      const data = resp.request.responseText.trimEnd()
      this.logsId = id
      const logs = data ? data.split('\n').map(line => JSON.parse(line)) : ''
      this.$set(this.fetchedLogs, id, logs)
    },
    toggleLogs (id) {
      const open = !this.openLogs[id]
      this.$set(this.openLogs, id, open)
      if (open && !this.fetchedLogs[id]) {
        this.loadLog(id)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.v-application.desktop .task-page {
  grid-column: 2 / 3;
  grid-row: 1 / 2;
  // overflow: auto;
  min-height: 0;
  max-height: 100%;

  .history {
    overflow: auto;
    .v-toolbar {
      position: sticky;
      top: 0;
      z-index: 1;
    }
  }
}
.v-card {
  max-width: 720px;
}
.v-list-item, .logs {
  border-bottom: 1px solid #eee;
  &:last-child {
    border-color: transparent;
  }
}
.logs {
  font-size: 13px;
  @media (max-width: 800px) {
    font-size: 11px;
  }
  // border-bottom: 1px solid #eee;
}
.id {
  min-width: 30px;
}
.history-list {
  display: grid;
  grid-template-columns: minmax(50px, auto) 1fr minmax(40px, auto) minmax(40px, auto);
  align-items: center;
  > div {
    display: contents;
  }
  .v-progress-circular {
    justify-self: center;
  }
}
</style>
