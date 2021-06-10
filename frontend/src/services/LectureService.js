import { baseUrl } from './HttpService.js'

export async function getLectureToday(token) {
  const rawResponse = await fetch(`${baseUrl}/lectures/today`, {
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

export async function getLecturesPast(token, moduleId) {
  const rawResponse = await fetch(`${baseUrl}/lectures/past?module=${moduleId}`, {
    method: 'GET',
    headers: {
      'Authorization': `${token}`,
    }
  })

  if (rawResponse.status == 204) {
    return []
  } else if(!rawResponse.ok) {
    throw rawResponse.status
  }
  
  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}

export async function completeLecture(token, lectureId) {
  const rawResponse = await fetch(`${baseUrl}/lectures/complete?id=${lectureId}`, {
    method: 'POST',
    headers: {
      'Authorization': `${token}`,
    }
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
}

export async function getLectureFlashcards(token, lectureId) {
  const rawResponse = await fetch(`${baseUrl}/lectures/flashcards?id=${lectureId}`, {
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
