import { baseUrl } from './HttpService.js'

export async function loginEmail(email) {
  const rawResponse = await fetch(`${baseUrl}/login/email?email=${email}`, {
    method: 'POST',
  })

  console.log(rawResponse)
  if(!rawResponse.ok) {
    throw rawResponse.status
  }
}

export async function verifyOtp(email, otp) {
  console.log(baseUrl)
  const rawResponse = await fetch(`${baseUrl}/verifyOtp?email=${email}&code=${otp}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    }
  })

  if(!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}
