import { baseUrl } from './HttpService.js'

export async function getSelf(token) {
  const rawResponse = await fetch(`${baseUrl}/self`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })
 
  if(!rawResponse.ok) {
    throw rawResponse.status
  }

  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}
