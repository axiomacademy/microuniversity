import { baseUrl } from './HttpService'

export async loginLearner(username, password) {
  // Create json
  let req = JSON.stringify({
    username: username,
    password: password
  });

  console.log(req)

  const rawResponse = await fetch(`${baseUrl}/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: req
  });

  console.log(rawResponse);

  if(!rawResponse.ok) {
    throw rawResponse.status;
  }

  // Return the json which is a JWT and the permission
  return await rawResponse.json();
}
