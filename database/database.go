package database

import (
	"GolangGraphQL/graph/model"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

var connectionString string = "mongodb://127.0.0.1:27017/GolangGraphQL"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("Error creating MongoDB client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel() // releases resources if Connect completes before timeout elapses
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	err = client.Ping(ctx, readpref.Primary()) // Ping MongoDB
	if err != nil {
		log.Fatalf("Error pinging MongoDB server: %v", err)
	}
	return &DB{client: client}
}

func (db *DB) GetJob(id string) *model.JobListing {
	jobCollection := db.client.Database("GolangGraphQL").Collection("jobs") // Get collection

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListing model.JobListing
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	err := jobCollection.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		log.Fatalf("Error getting job: %v", err)
	}
	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollection := db.client.Database("GolangGraphQL").Collection("jobs") // Get collection

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // releases resources if Connect completes before timeout elapses

	var jobListings []*model.JobListing

	cursor, err := jobCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatalf("Error getting jobs: %v", err)
	}
	if err = cursor.All(context.TODO(), &jobListings); err != nil {
		log.Fatalf("Error getting job listingd: %v", err)
	}

	return jobListings
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("GolangGraphQL").Collection("jobs") // Get collection

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // releases resources if Connect completes before timeout elapsed

	inserted, err := jobCollection.InsertOne(ctx, bson.M{
		"title":       jobInfo.Title,
		"description": jobInfo.Description,
		"url":         jobInfo.URL,
		"company":     jobInfo.Company,
	})
	if err != nil {
		log.Fatalf("Error Creating job %v", err)
	}

	// get the id which was inserted
	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()

	returnJobListing := model.JobListing{
		ID:          insertedID,
		Title:       jobInfo.Title,
		Company:     jobInfo.Company,
		Description: jobInfo.Description,
		URL:         jobInfo.URL,
	}

	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("GolangGraphQL").Collection("jobs") // Get collection

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // releases resources if Connect completes before timeout elapses

	updateJobInfo := bson.M{}
	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Company != nil {
		updateJobInfo["company"] = jobInfo.Company
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}

	var jobListing model.JobListing

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}            // filter by id
	update := bson.M{"$set": updateJobInfo} // update the job listing
	results := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	if err := results.Decode(&jobListing); err != nil {
		log.Fatalf("Error updating job: %v", err)
	}
	return &jobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	jobCollection := db.client.Database("GolangGraphQL").Collection("jobs") // Get collection

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // releases resources if Connect completes before timeout elapses

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	_, err := jobCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatalf("Error deleting job: %v", err)
	}

	return &model.DeleteJobResponse{DeleteJobID: jobId}
}
