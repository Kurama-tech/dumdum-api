package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name  string `json:"name"`
	Company string `json:"company"`
	Avatar string `json:"avatar"`
	Images []string `bson:"images" json:"images"`
	Proprietor string `json:"proprietor"`
	Contact string `json:"contact"`
	UserId string `json:"uid"`
	Status string `json:"status"`
	Location string `json:"location"`
	BusniessCategory string `json:"busniess_category"`
	BusniessType string `json:"busniess_type"`
}

type UserGet struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name  string `json:"name"`
	Company string `json:"company"`
	Avatar string `json:"avatar"`
	Images []string `bson:"images" json:"images"`
	Proprietor string `json:"proprietor"`
	Contact string `json:"contact"`
	UserId string `json:"uid"`
	Status string `json:"status"`
	Location string `json:"location"`
	BusniessCategory string `json:"busniess_category"`
	BusniessType string `json:"busniess_type"`
}


const Database = "jwc"
const minioURL = "167.71.233.124:9000"

func getEnv(Environment string) (string, error) {
	variable := os.Getenv(Environment)
	if variable == "" {
		fmt.Println(Environment + ` Environment variable is not set`)
		return "", errors.New("env Not Set Properly")
	} else {
		fmt.Printf(Environment + " variable value is: %s\n", variable)
		return variable, nil
	}
}
func main() {
	// MongoDB client options
	
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // All origins
		AllowedMethods: []string{"POST", "GET", "PUT", "DELETE"}, // Allowing only get, just an example
		AllowedHeaders: []string{"Set-Cookie", "Content-Type"},
		ExposedHeaders: []string{"Set-Cookie"},
		AllowCredentials: true,
		Debug: true,
	})
	

	// Get the value of the "ENV_VAR_NAME" environment variable
	mongoURL, err := getEnv("MONGO_URL")
	if err != nil {
		os.Exit(1)
	}

	minioKey, err:= getEnv("MINIO_KEY")
	if err != nil {
		os.Exit(1)
	}
	minioSecret, err:= getEnv("MINIO_SECRET")
	if err != nil {
		os.Exit(1)
	}

	
	clientOptions := options.Client().ApplyURI(mongoURL)
	
	// Create a new MongoDB client
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	
	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	minioClient, err := minio.New(minioURL, minioKey, minioSecret, false)
    if err != nil {
        log.Fatalln(err)
    }
	
	// Create a new router using Gorilla Mux
	router := mux.NewRouter()

	
	// Define a POST route to add an item to a collection
	router.HandleFunc("/api/user", addItem(client)).Methods("POST")

	// Get All Users
	router.HandleFunc("/api/users", getItems(client)).Methods("GET")

	// get disabled users
	router.HandleFunc("/api/users/disabled", getDisabledItems(client)).Methods("GET")

	// get user based on id
	router.HandleFunc("/api/users/{id}", getItem(client)).Methods("GET")
	
	// Define a DELETE route to delete an user from a collection
	router.HandleFunc("/api/users/delete/{id}", deleteItem(client)).Methods("DELETE")

	//disable a user based on id
	router.HandleFunc("/api/users/disable/{id}", disableItem(client)).Methods("DELETE")

	// enable a user based on id
	router.HandleFunc("/api/users/enable/{id}", enableItem(client)).Methods("GET")

	// upload an image/images to minio bucket and get url or list of urls
	router.HandleFunc("/api/upload", Upload(minioClient)).Methods("POST")
	
	// Define a PUT route to edit an item in a collection
	router.HandleFunc("/api/users/{id}", editItem(client)).Methods("PUT")


	
	// Start the HTTP server
	log.Println("Starting HTTP server...")
	err = http.ListenAndServe(":8002", c.Handler(router))
	if err != nil {
		
		log.Fatal(err)
	}
}

func Upload(minioClient *minio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        // Parse the multipart form.
        err := r.ParseMultipartForm(32 << 20)
        if err != nil {
			fmt.Println(err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Get the file headers from the form.
        files := r.MultipartForm.File["files"]

		var ImagePaths []string 

        // Loop through the files and upload them to Minio.
        for _, fileHeader := range files {
            // Open the file.
            file, err := fileHeader.Open()
            if err != nil {
				fmt.Println(err.Error())
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            defer file.Close()

            // Get the file name and extension.
            filename := fileHeader.Filename
            extension := filepath.Ext(filename)

            // Generate a unique file name with the original extension.
            newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
			newPath := "http://"+minioURL + "/dumdum/" + newFilename
			ImagePaths = append(ImagePaths, newPath)

            // Upload the file to Minio.
            _, err = minioClient.PutObject("jwc", newFilename, file, fileHeader.Size, minio.PutObjectOptions{
				ContentType: "image/png",
			})
            if err != nil {
				fmt.Println(err.Error())
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }

		data, err := json.Marshal(ImagePaths)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to application/json
		//w.Header().Set("Content-Type", "application/json")
		// Write the JSON data to the response
		
        // Send a success response.
        w.WriteHeader(http.StatusOK)
		w.Write(data)
    }
}


// addItem inserts a new item into the "items" collection in MongoDB
func addItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item User
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

// deleteItem deletes an item from the "items" collection in MongoDB
func deleteItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]
		
		// Delete the item from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func editItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		
		
		// Parse the request body into an Item struct
		var item UserGet
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)

		// tables := item.TableAttached
		
		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		filter := bson.M{"_id": item.ID}
		update := bson.M{"$set": bson.M{"name": item.Name, "company":item.Company,"avatar": item.Avatar,"proprietor": item.Proprietor, "status": item.Status, "images": item.Images, "location": item.Location }}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func disableItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		
		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "disabled"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func enableItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		
		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "active"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getItems(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		// Decode the cursor results into a slice of Item structs
		var items []UserGet
		for cursor.Next(context.Background()) {
			var item UserGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getDisabledItems(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		cursor, err := collection.Find(context.Background(), bson.M{"status": "disabled"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		// Decode the cursor results into a slice of Item structs
		var items []UserGet
		for cursor.Next(context.Background()) {
			var item UserGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("users")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cursor, err := collection.Find(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		
		// Decode the cursor results into a slice of Item structs
		var items []UserGet
		for cursor.Next(context.Background()) {
			var item UserGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// searchItems retrieves all items from the "items" collection in MongoDB that match a search key
func searchItems(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the search key parameter from the request URL
		vars := mux.Vars(r)
		key := vars["key"]
		
		// Search for items in the "items" collection in MongoDB that match the search key
		collection := client.Database(Database).Collection("products")
		cursor, err := collection.Find(context.Background(), bson.M{"name": primitive.Regex{Pattern: key, Options: "i"}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		// Decode the cursor results into a slice of Item structs
		var items []UserGet
		for cursor.Next(context.Background()) {
			var item UserGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		
		// Send the list of matching items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
