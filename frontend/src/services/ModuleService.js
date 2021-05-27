import { baseUrl } from './HttpService.js'
 
export async function getModules() {
  const rawResponse = await fetch(`${baseUrl}/modules`, {
    method: 'GET',
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  }
  
  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}
