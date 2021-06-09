import { baseUrl } from './HttpService.js'

export async function getSelfCohorts(token) {
  const rawResponse = await fetch(`${baseUrl}/cohorts`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })
 
  if(rawResponse.status == 204) {
    return null
  } else if(!rawResponse.ok) {
    throw rawResponse.status
  }

  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}

export async function getAvailableCohorts(token, moduleId) {
  const rawResponse = await fetch(`${baseUrl}/cohorts/available?module=${moduleId}`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })

  if(rawResponse.status == 204) {
    return null
  } else if(!rawResponse.ok) {
    throw rawResponse.status
  }

  let json = await rawResponse.json()
  if (json == null) {
    return []
  } else {
    return json
  }
}

// Your applied/accepted cohort for a module
export async function getModuleCohort(token, moduleId) {
  const rawResponse = await fetch(`${baseUrl}/cohort/self?module=${moduleId}`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })

  if(rawResponse.status == 204) {
    return null
  } else if(!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}

export async function joinCohort(token, cohortId) {
  const rawResponse = await fetch(`${baseUrl}/cohort/join?cohort=${cohortId}`, {
    method: 'POST',
    headers: {
      'Authorization': `${token}`,
    }
  })
 
  if(!rawResponse.ok) {
    throw rawResponse.status
  }
}

export async function leaveModuleCohort(token, moduleId) {
  const rawResponse = await fetch(`${baseUrl}/cohort/leave?module=${moduleId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `${token}`,
    }
  })
 
  if(!rawResponse.ok) {
    throw rawResponse.status
  }
}
