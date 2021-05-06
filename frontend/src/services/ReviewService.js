import { baseUrl } from './HttpService.js'

export async function getDailyReview(token) {
  const rawResponse = await fetch(`${baseUrl}/review`, {
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

export async function passFlashcard(token, flashcardId) {
  const rawResponse = await fetch(`${baseUrl}/flashcard/pass?id=${flashcardId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  }) 
  
  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
}

export async function failFlashcard(token, flashcardId) {
  const rawResponse = await fetch(`${baseUrl}/flashcard/fail?id=${flashcardId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  }) 
  
  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
}

export async function completeReview(token) {
  const rawResponse = await fetch(`${baseUrl}/review/complete`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  }) 
  
  if(!rawResponse.ok) {
    throw rawResponse.status
  } 
}
