import { graphqlUrl } from './HttpService.js'

const getCoreDataQuery = `
  query GetHomeData($email: String!) {
    getLearner(email: $email) {
      firstName,
      lastName,
      challenges {
        status,
        challenge {
          id,
          title,
          description
        }
      },
      unlockedTutorials {
        id,
        title,
        description
      },
      activeCohort {
        status,
        tutorial {
          title,
          description,
        },
      },
      masteredTopics {
        id,
        name,
      },
      currentPlanet {
        id,
        minedKnowledge,
        completed,
        planet {
          name,
          starSystem {
            name,
          }
        },
      },
      energy,
      coins
    }
  }`

const getDailyReviewQuery = `
  query GetDailyReview($email: String!) {
    getLearner(email: $email) {
      dailyReview {
        id,
        reviewCard {
          topText,
          bottomText
        }
      }
    }
  }`

const getRecommendedLecturesQuery = `
  query GetRecommendedLectures($email: String!) {
    getLearner(email: $email) {
      recommendedLectures {
        title,
      }
    }
  }`
          
export async function getCoreData(token, email) {
  const rawResponse = await fetch(graphqlUrl, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({
      query: getCoreDataQuery,
      variables: { email }
    })
  })

  if (!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}

export async function getDailyReview(token, email) {
  const rawResponse = await fetch(graphqlUrl, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({
      query: getDailyReviewQuery,
      variables: { email }
    })
  })

  if (!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}

export async function getRecommendedLectures(token, email) {
  const rawResponse = await fetch(graphqlUrl, {
    method: "POST",
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': token,
    },
    body: JSON.stringify({
      query: getRecommendedLecturesQuery,
      variables: { email }
    })
  })

  if (!rawResponse.ok) {
    throw rawResponse.status
  }

  return await rawResponse.json()
}
