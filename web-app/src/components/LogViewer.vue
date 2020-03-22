<script>
export default {
  name: 'LogViewer',
  props: {
    logs: Array
  },
  cumputed: {
    stdout () {
      return this.logs.filter(log => log.stream === 'stdout').map(log => log.log)
    }
  },
  render (h) {
    const logs = this.logs.map(log => h('span', { class: log.stream }, log.log))
    return h('pre', logs)
  }
}
</script>

<style lang="scss" scoped>
pre {
  white-space: pre-wrap;
  background-color: #fafafa;
}
.stderr {
  color: red;
}
</style>
