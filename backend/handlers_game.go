package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Go to a nearby planet (in the same starsystem)
// Requires: planetId, energy, learnerId
// 1. Check energy requirements
// 2. Check that the planet is nearby
// 3. Set currentPlanet to the new planet, and subtract energy costs
func (s *server) handleGotoPlanet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		luid := r.Header.Get("X-Uid-Claim")
		energy, _ := strconv.Atoi(r.Header.Get("X-Energy-Claim"))

		query := r.URL.Query()
		planetId := query.Get("planetId")

		if planetId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := energy - PLANET_ENERGY_DEPLETION; e >= 0 {
			energy = e
		} else {
			fmt.Println("Not enough energy")
			http.Error(w, "Not enough energy", http.StatusBadRequest)
			return
		}

		// Make sure the planet is visitable
		const checkPlanetNearby = `
		query checkPlanetNearby($planetId: string, $learnerId: string) {
			checkPlanetNearby(func: uid($learnerId)) @cascade {
				Learner.currentPlanet @cascade {
					LearnerPlanet.planet @cascade {
						Planet.starSystem @cascade {
							StarSystem.planets @filter(uid($planetId)) {
								uid
							}
						}
					}
				}
			}
		}

		query checkPlanetVisited($planetId: string, $learnerId: string) {
			checkPlanetVisited(func: type(LearnerPlanet)) @filter(uid_in(planet, $planetId) AND uid_in(learner, $learnerId)) {
				uid
			}
		}
	`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkPlanetNearby, map[string]string{
			"$planetId":  planetId,
			"$learnerId": luid,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var d struct {
			CheckPlanetNearby  []Learner
			CheckPlanetVisited []LearnerPlanet
		}

		if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If there's nothing means it's not nearby
		if len(d.CheckPlanetNearby) == 0 {
			fmt.Println("Planet not nearby")
			http.Error(w, "Planet not nearby", http.StatusBadRequest)
			return
		}

		// Check if planet visited
		var l Learner
		if len(d.CheckPlanetVisited) == 0 {
			// Planet not visited
			l = Learner{
				Uid:    luid,
				Energy: energy,
				CurrentPlanet: LearnerPlanet{
					Planet:         Planet{Uid: planetId},
					Learner:        &Learner{Uid: luid},
					MinedKnowledge: 0,
					Completed:      false,
				},
			}
		} else {
			l = Learner{
				Uid:    luid,
				Energy: energy,
				CurrentPlanet: LearnerPlanet{
					Uid: d.CheckPlanetVisited[0].Uid,
				},
			}
		}

		pl, err := json.Marshal(l)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu := &api.Mutation{
			SetJson: pl,
		}

		_, err = txn.Mutate(r.Context(), mu)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			err = txn.Commit(r.Context())
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		*/

		return
	}
}

// Go to nearby Starsystem
// Requires: starSystemId, planetId, energy, learnerId
// 1. Check energy requirements
// 2. Check that starsystem is nearby and planet is inside starsystem
// 3. Set currentPlanet to the new planet and subtract energy costs

func (s *server) handleGotoStarsystem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		luid := r.Header.Get("X-Uid-Claim")
		energy, _ := strconv.Atoi(r.Header.Get("X-Energy-Claim"))

		query := r.URL.Query()
		planetId := query.Get("planetId")
		systemId := query.Get("systemId")

		if planetId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		if systemId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := energy - STARSYSTEM_ENERGY_DEPLETION; e >= 0 {
			energy = e
		} else {
			fmt.Println("Not enough energy")
			http.Error(w, "Not enough energy", http.StatusBadRequest)
			return
		}

		// Make sure the planet is visitable
		const checkSystemNearby = `
		query checkPlanetNearby($systemId: string, $planetId: string, $learnerId: string) {
			checkPlanetNearby(func: uid($learnerId)) @cascade {
				Learner.currentPlanet @cascade {
					LearnerPlanet.planet {
						Planet.starSystem @cascade {
							StarSystem.nearbySystems @cascade {
								uid
								StarSystem.name
								StarSystem.planets @filter(uid($planetId)) {
									uid
								}
							}
						}
					}
				}
			}
		}
		query checkPlanetVisited($planetId: string, $learnerId: string) {
			checkPlanetVisited(func: type(LearnerPlanet)) @filter(uid_in(planet, $planetId) AND uid_in(learner, $learnerId)) {
				uid
			}
		}
	`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkSystemNearby, map[string]string{
			"$systemId":  systemId,
			"$planetId":  planetId,
			"$learnerId": luid,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var d struct {
			CheckSystemNearby  []Learner
			CheckPlanetVisited []LearnerPlanet
		}

		if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If there's nothing means it's not nearby
		if len(d.CheckSystemNearby) == 0 {
			fmt.Println("System not nearby")
			http.Error(w, "System not nearby", http.StatusBadRequest)
			return
		}

		// TODO: generate planets and systems programatically

		var l Learner
		if len(d.CheckPlanetVisited) == 0 {
			// Planet not visited
			l = Learner{
				Uid:    luid,
				Energy: energy,
				CurrentPlanet: LearnerPlanet{
					Planet:         Planet{Uid: planetId},
					Learner:        &Learner{Uid: luid},
					MinedKnowledge: 0,
					Completed:      false,
				},
			}
		} else {
			l = Learner{
				Uid:           luid,
				Energy:        energy,
				CurrentPlanet: LearnerPlanet{Uid: d.CheckPlanetVisited[0].Uid},
			}
		}

		pl, err := json.Marshal(l)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu := &api.Mutation{
			SetJson: pl,
		}

		_, err = txn.Mutate(r.Context(), mu)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			err = txn.Commit(r.Context())
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		*/
		return
	}
}
