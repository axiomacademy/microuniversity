package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"google.golang.org/grpc"
)

type testDb struct {
	dg *dgo.Dgraph
}

func newTestDb() *testDb {
	// Getting all the environmental variables
	DB_URL := os.Getenv("DB_URL")
	checkEnvVariable(DB_URL)

	// Initialise the dgraph database
	conn, err := grpc.Dial(DB_URL, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Sprintf("Could not connect to grpc endpoint. Got error %v", err.Error()))
	}

	dg := dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)

	db := &testDb{
		dg: dg,
	}

	db.Populate()
	return db
}

// Populate database with the test data
func (db *testDb) Populate() {
	db.loadFile("./testdata/C-0001.rdf")
	db.loadFile("./testdata/subjects.rdf")
	db.loadFile("./testdata/universe.rdf")
}

func (db *testDb) Teardown() {
	op := &api.Operation{
		DropAll: true,
	}

	if err := db.dg.Alter(context.Background(), op); err != nil {
		panic(fmt.Sprintf("Could not cleanup database. Got error %v", err.Error()))
	}
}

// Load the triples into dgraph
// Requires: triples, dgraph client
func (db *testDb) addTriples(triples string) error {
	txn := db.dg.NewTxn()
	ctx := context.Background()
	defer txn.Discard(ctx)

	_, err := txn.Mutate(ctx, &api.Mutation{
		SetNquads: []byte(triples),
		CommitNow: true,
	})

	return err
}

// Load a file containing triples into dgraph
// Requires: data file url, dgraph client
func (db *testDb) loadFile(data string) {
	rdf, err := ioutil.ReadFile(data)
	if err != nil {
		panic(fmt.Sprintf("Could not read test data file. Got error %v", err.Error()))
	}

	if err = db.addTriples(string(rdf)); err != nil {
		panic(fmt.Sprintf("Could not load test data file. Got error %v", err.Error()))
	}
}
