export default $axios => ({
  login: function(username, password) {
    return $axios.$post('/subscribe', {
      username: username,
      password: password
    })
  },
  start: function(token) {
    return $axios.$post(
      '/control',
      {
        action: 'start'
      },
      {
        headers: {
          Authorization: `Bearer ${token}`
        }
      }
    )
  },
  stop: function(token) {
    return $axios.$post(
      '/control',
      {
        action: 'stop'
      },
      {
        headers: {
          Authorization: `Bearer ${token}`
        }
      }
    )
  }
})
