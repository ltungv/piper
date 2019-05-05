import https from 'https'
import fs from 'fs'

const HOST = '127.0.0.1'
const PORT = '4433'

const httpsAgent = new https.Agent({
  rejectUnauthorized: false,
  ca: fs.readFileSync('../../keys/certs/pub/cacert.pem'),
  cert: fs.readFileSync('../../keys/certs/pub/clientcert.pem'),
  key: fs.readFileSync('../../keys/certs/priv/clientkey.pem')
})

const url = `https://${HOST}:${PORT}`

export default $axios => ({
  login(username, password) {
    return $axios.$post(
      `${url}/subscribe`,
      {
        username: username,
        password: password
      },
      { httpsAgent }
    )
  }
})
