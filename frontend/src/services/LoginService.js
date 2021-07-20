import { graphqlUrl } from './HttpService.js'

const getSelfQuery = `
  query GetSelf($email: String!) {
    getLearner(email: $email) {
      email,
      firstName,
      lastName
    }
  }
`

export async function getSelf(token, email) {
  const rawResponse = await fetch(graphqlUrl, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({
      query: getSelfQuery,
      variables: { email }
    })
  })

  if (!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}

const createSelfMutation = `
  mutation CreateSelf($email: String!, $firstName: String!, $lastName: String!, $timezone: String!) {
    addLearner(input: {
      email: $email,
      firstName: $firstName,
      lastName: $lastName,
      timezone: $timezone,
      completedLectures: [],
      cards: [],
      challenges: []
    }) {
      learner {
        email
      }
    }
  }
`

export async function createSelf(token, email, firstName, lastName, timezone) {
  const rawResponse = await fetch(graphqlUrl, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({
      query: createSelfMutation,
      variables: { email, firstName, lastName, timezone }
    })
  })

  if (!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}
