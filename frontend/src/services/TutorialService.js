import { baseUrl } from './HttpService.js'

export async function getUpcomingTutorials(token, moduleId) {
  const rawResponse = await fetch(`${baseUrl}/tutorials?module=${moduleId}`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })

  if(rawResponse.status == 204) {
    return []
  } else if(!rawResponse.ok) {
    throw rawResponse.status
  }
  
  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}
