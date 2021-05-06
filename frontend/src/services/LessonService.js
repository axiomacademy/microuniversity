import { baseUrl } from './HttpService.js'

export async function getLessonToday(token) {
  const rawResponse = await fetch(`${baseUrl}/lessons/today`, {
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

export async function getLessonsPast(token) {
  const rawResponse = await fetch(`${baseUrl}/lessons/past`, {
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

export async function completeLesson(token, lessonId) {
  const rawResponse = await fetch(`${baseUrl}/lessons/complete?id=${lessonId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
}

export async function getLessonFlashcards(token, lessonId) {
  const rawResponse = await fetch(`${baseUrl}/lessons/flashcards?id=${lessonId}`, {
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
