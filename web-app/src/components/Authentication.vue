<template>
  <v-dialog
    v-model="visible"
    max-width="500"
    persistent
  >
    <v-card>
      <v-card-title>Authentication</v-card-title>
      <v-card-text>
        <v-text-field
          label="Password"
          type="password"
          :error="error"
          v-model="password"
          @keydown.enter="login"
        />
      </v-card-text>
      <v-card-actions class="mx-4 pb-3">
        <v-spacer/>
        <v-btn
          color="secondary"
          :disabled="!password"
          @click="login"
        >
          <span>Ok</span>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  data () {
    return {
      error: false,
      visible: false,
      password: ''
    }
  },
  created () {
    this.$http.setAuthHandler(() => {
      this.visible = true
      return new Promise((resolve, reject) => {
        this._promise = { resolve, reject }
      })
    })
  },
  methods: {
    login () {
      this.$http.post('/api/auth/login', { password: this.password })
        .then(() => {
          this.error = false
          this.visible = false
          this._promise && this._promise.resolve()
        })
        .catch(() => {
          this.error = true
        })
    }
  }
}
</script>
