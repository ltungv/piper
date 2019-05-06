const HOST = '127.0.0.1'
const PORT = '4433'

const url = `https://${HOST}:${PORT}`

export default $axios => ({
  login: function(username, password) {
    return $axios.$post(`${url}/subscribe`, {
      username: username,
      password: password
    })
  },
  start: function(token) {
    return $axios.$post(
      `${url}/control`,
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
      `${url}/control`,
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
