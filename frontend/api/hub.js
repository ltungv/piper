const HOST = '192.168.1.100'
const PORT = '4433'

const URL = `https://${HOST}:${PORT}`

export default $axios => ({
  login: function(username, password) {
    return $axios.$post(`${URL}/subscribe`, {
      username: username,
      password: password
    })
  },
  start: function(token) {
    return $axios.$post(
      `${URL}/control`,
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
      `${URL}/control`,
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
