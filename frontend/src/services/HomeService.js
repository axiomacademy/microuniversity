import { graphqlUrl } from './HttpService.js'

const getCoreDataQuery = `
  query GetHomeData($email: String!) {
    getLearner(email: $email) {
      firstName,
      lastName,
      challenges {
        status,
        challenge {
          title,
          description
        }
      },
      unlockedTutorials {
        title,
        description
      },
      activeCohorts {
        tutorial {
          title,
          description,
        },
        status
      },
      masteredTopics {
        name
      },
      currentPlanet {
        name,
        totalKnowledge,
        minedKnowledge,
        starSystem {
          name,
        }
      },
      energy,
      coin
    }
  }`

const getDailyReviewQuery = `
  query GetHomeData($email: String!) {
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
  query GetHomeData($email: String!) {
    getLearner(email: $email) {
      recommendedLectures {
        title,
        description,
        videoLink
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
