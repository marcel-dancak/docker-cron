import axios from 'axios'

const http = axios.create({
  baseURL: '',
  withCredentials: true
})

let authInterceptor = null
http.setAuthHandler = function (handleAuth) {
  if (authInterceptor !== null) {
    http.interceptors.response.eject(authInterceptor)
  }
  authInterceptor = http.interceptors.response.use(null, error => {
    if (error.config && error.response && error.response.status === 403) {
      return handleAuth().then(() => http.request(error.config))
    }
    return Promise.reject(error)
  })
}

export default http
