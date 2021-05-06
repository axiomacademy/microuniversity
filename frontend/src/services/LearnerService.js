import { baseUrl } from './HttpService.js'

export async function loginLearner(username, password) {
  // Create json
  let req = JSON.stringify({
    username: username,
    password: password
  })

  const rawResponse = await fetch(`${baseUrl}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: req
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  }

  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}

export async function getSelf(token) {
  const rawResponse = await fetch(`${baseUrl}/self`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  })
 
  if(!rawResponse.ok) {
    throw rawResponse.status
  }

  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}
