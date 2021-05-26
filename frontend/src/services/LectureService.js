import { baseUrl } from './HttpService.js'

export async function getLectureToday(token) {
  const rawResponse = await fetch(`${baseUrl}/lectures/today`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
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

export async function getLecturesPast(token) {
  const rawResponse = await fetch(`${baseUrl}/lectures/past`, {
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

export async function completeLecture(token, lectureId) {
  const rawResponse = await fetch(`${baseUrl}/lectures/complete?id=${lectureId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
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
      'Authorization': `Bearer ${token}`,
    }
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
  
  // Return the json which is a JWT and the permission
  return await rawResponse.json()
}