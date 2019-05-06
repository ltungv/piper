<template>
  <section id="app" class="hero is-light is-fullheight is-bold">
    <div class="hero-body">
      <div class="container has-text-centered">
        <form @submit.prevent="adminLogin">
          <input v-model="username" type="text" />
          <input v-model="password" type="password" />
          <button type="submit">Submit</button>
        </form>
      </div>
    </div>
  </section>
</template>

<script>
const Cookie = process.client ? require('js-cookie') : undefined

export default {
  middleware: 'notAuthenticated',
  components: {},
  data: function() {
    return {
      username: '',
      password: ''
    }
  },
  methods: {
    adminLogin: async function() {
      try {
        const resp = await this.$hub.login(this.username, this.password)
        const { token } = resp
        this.$store.commit('setToken', token)
        Cookie.set('RoboconToken', token, { expires: 2, path: '' })
        this.$router.push('/')
      } catch (e) {
        console.log(e)
      }
    }
  }
}
</script>

<style lang="scss"></style>
