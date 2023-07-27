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
	"sort"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
)

type User struct {
	Name  string `json:"name"`
	Company string `json:"company"`
	Avatar string `json:"avatar"`
	Images []string `bson:"images" json:"images"`
	Proprietor string `json:"proprietor"`
	Contact string `json:"contact"`
	UserId string `json:"userid"`
	Status string `json:"status"`
	Location string `json:"location"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	BusniessCategory string `json:"busniess_category"`
	BusniessType string `json:"busniess_type"`
}

type UsersDevices struct {
	UserId string `json:"userid"`
	DeviceId string `json:"deviceid"`
}

type UsersDevicesGet struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	UserId string `json:"userid"`
	DeviceId string `json:"deviceid"`
}

type Requirements struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Priority string `json:"priority"`
	Status string `json:"status"`
	UserId string `json:"userid"`
	Remarks string `json:"remarks"`
	Timestamp time.Time `bson:"timestamp"`
	LastUpdate time.Time `json:"lastUpdate"`
}

type RequirementsGet struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Title string `json:"title"`
	Description string `json:"description"`
	Priority string `json:"priority"`
	Status string `json:"status"`
	UserId string `json:"userid"`
	Remarks string `json:"remarks"`
	Timestamp time.Time `bson:"timestamp"`
	LastUpdate time.Time `json:"lastUpdate"`
}



type Conversation struct {
	Started bool `json:"started"`
	ByUserId string `json:"byuserid"`
	WithUserId string `json:"withuserid"`
}

type ConversationGet struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Started bool `json:"started"`
	ByUserId string `json:"byuserid"`
	WithUserId string `json:"withuserid"`
}

type Conversations struct {
	Conversation ConversationGet `json:"conversation"`
	ByUser      UserGet             `json:"byuser"`
	WithUser    UserGet             `json:"withuser"`
	DisplayUser UserGet				`json:"displayuser"`
}

type HistoryMessage struct {
	Title string `bson:"title"`
    Message   string    `bson:"message"`
    Timestamp time.Time `bson:"timestamp"`
}

type Contacted struct {
	UserId string `json:"userid"`
	Name  string `json:"name"`
}

type Connected struct {
	UserId string `json:"userid"`
	Name  string `json:"name"`
}

type Follow struct {
	UserId string `json:"userid"`
	Contacted []Contacted `bson:"contacted"`
	Connected []Connected `bson:"connected"`
}
type History struct {
    UserId    string    `bson:"userid"`
    Messages  []HistoryMessage    `bson:"messages"`
}

type DeleteImage struct {
	URL string `bson:"url"`
}

type Contact struct{
	UserId string `bson:"userid"`
	MyName string `bson:"myname"`
	Name string `bson:"name"`
}


type Connections struct {
	UserId    string    `bson:"userid"`
	Contacted int `bson:"contacted"`
	Connections int `bson:"connections"`
}

type UserGet struct {
	ID    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name  string `json:"name"`
	Company string `json:"company"`
	Avatar string `json:"avatar"`
	Images []string `bson:"images" json:"images"`
	Proprietor string `json:"proprietor"`
	Contact string `json:"contact"`
	UserId string `json:"userid"`
	Status string `json:"status"`
	Location string `json:"location"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	BusniessCategory string `json:"busniess_category"`
	BusniessType string `json:"busniess_type"`
}


const Database = "dumdum"

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

	minioURL, err := getEnv("MINIO_URL")
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

	minioClient, err := minio.New(minioURL, &minio.Options{
		Creds:  credentials.NewStaticV4(minioKey, minioSecret, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", minioClient)

	opt := option.WithCredentialsFile("service-account.json")
    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        log.Fatalf("Error initializing Firebase app: %v\n", err)
    }

    FCMclient, err := app.Messaging(context.Background())
    if err != nil {
        log.Fatalf("Error initializing Firebase Messaging client: %v\n", err)
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
	router.HandleFunc("/api/upload", Upload(minioClient, minioURL)).Methods("POST")
	
	// Define a PUT route to edit an item in a collection
	router.HandleFunc("/api/users/{id}", editItem(client)).Methods("PUT")

	router.HandleFunc("/api/upload/image/{id}", UploadUserImages(minioClient, client, minioURL)).Methods("POST")

	router.HandleFunc("/api/add/history/{id}", AddHistoryMessage(client)).Methods("POST")

	router.HandleFunc("/api/list/history/{id}", ListHistory(client)).Methods("GET")
	router.HandleFunc("/api/list/connections/{id}", ListConnections(client)).Methods("GET")

	router.HandleFunc("/api/contact/{id}", IncrementContacted(client)).Methods("PUT")

	router.HandleFunc("/api/delete/image/{id}", DeleteUserImages(minioClient, client)).Methods("DELETE")

	router.HandleFunc("/api/search/{key}", searchItems(client)).Methods("GET")

	router.HandleFunc("/api/list/follow/{id}", getFollowItem(client)).Methods("GET")

	router.HandleFunc("/api/conversation/start", StartConversation(client)).Methods("POST")

	router.HandleFunc("/api/conversation/list/{id}", GetConversations(client)).Methods("GET")

	router.HandleFunc("/api/requirements/list/{id}", getRequirement(client)).Methods("GET")

	router.HandleFunc("/api/requirements", getRequirements(client)).Methods("GET")

	router.HandleFunc("/api/requirements", addRequirement(client, FCMclient)).Methods("POST")

	router.HandleFunc("/api/requirements/close/{id}", closeRequirement(client)).Methods("POST")

	router.HandleFunc("/api/requirements/update/{id}", editRequirement(client)).Methods("PUT")

	router.HandleFunc("/api/devices/add", addDevice(client)).Methods("POST")
	
	router.HandleFunc("/api/devices/list/{id}", getDevice(client)).Methods("GET")





	
	// Start the HTTP server
	log.Println("Starting HTTP server...")
	err = http.ListenAndServe(":8002", c.Handler(router))
	if err != nil {
		
		log.Fatal(err)
	}
}

func ListHistory(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("history")
		filter := bson.M{"userid": id}

		// Sort the messages by timestamp (newest first)
		sortOptions := options.Find().SetSort(bson.M{"messages.timestamp": -1})

		cursor, err := collection.Find(context.Background(), filter, sortOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of History structs
		var items []History
		for cursor.Next(context.Background()) {
			var item History

			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Sort the messages for this history item by timestamp (newest first)
			sort.Slice(item.Messages, func(i, j int) bool {
				return item.Messages[i].Timestamp.After(item.Messages[j].Timestamp)
			})

			items = append(items, item)
		}

		fmt.Println(items)

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func GetConversations(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		// Define the filter to match conversations where the id matches either ByUserId or WithUserId
		filter := bson.M{
			"$or": []bson.M{
				{"byuserid": id},
				{"withuserid": id},
			},
		}

		// Retrieve the conversations from the "conversation" collection in MongoDB
		collection := client.Database(Database).Collection("conversation")
		cur, err := collection.Find(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cur.Close(context.Background())

		// Iterate through the result cursor and collect the conversations
		var conversations []Conversations
		for cur.Next(context.Background()) {
			var conversation ConversationGet
			if err := cur.Decode(&conversation); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Retrieve the user information for the "ByUserId" and "WithUserId" values
			byUserID := conversation.ByUserId
			withUserID := conversation.WithUserId

			// Retrieve the user objects from the "users" collection
			usersCollection := client.Database(Database).Collection("users")

			byUser := UserGet{}
			err = usersCollection.FindOne(context.Background(), bson.M{"userid": byUserID}).Decode(&byUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			withUser := UserGet{}
			err = usersCollection.FindOne(context.Background(), bson.M{"userid": withUserID}).Decode(&withUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			displayUser := UserGet{}
			if(id == byUser.UserId){
				displayUser = withUser
			}else {
				displayUser = byUser
			}

			// Append conversation, byUser, and withUser to the result
			result := Conversations{
				Conversation: conversation,
				ByUser:       byUser,
				WithUser:     withUser,
				DisplayUser: displayUser,
				
			}

			conversations = append(conversations, result)
		}

		if err := cur.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the conversations as JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conversations)
	}
}


func ListConnections(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request)  {
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("connections")
		// oid, err := primitive.ObjectIDFromHex(id)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		cursor, err := collection.Find(context.Background(), bson.M{"userid": id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		
		// Decode the cursor results into a slice of Item structs
		var items []Connections
		for cursor.Next(context.Background()) {
			var item Connections

			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		fmt.Println(items)
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func IncrementContacted(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	id := vars["id"]

	var item Contact
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := client.Database(Database).Collection("connections")
    filter := bson.M{"userid": item.UserId}
    update := bson.M{"$inc": bson.M{"connections": 1}}

    _, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}



	collection = client.Database(Database).Collection("connections")
    filter = bson.M{"userid": id}
    update = bson.M{"$inc": bson.M{"contacted": 1}}

    _, err = collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updateTimeline := HistoryMessage{
		Title: "Contacted user",
		Message: "Contacted user"+ item.Name,
		Timestamp: time.Now(),

	}

	contacted := Connected{
		Name: item.Name,
		UserId: item.UserId,
	}

	collection = client.Database(Database).Collection("history")
	filter = bson.M{"userid": id}
	update = bson.M{"$push": bson.M{"messages": updateTimeline}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	collection = client.Database(Database).Collection("follow")
	filter = bson.M{"userid": id}
	update = bson.M{"$push": bson.M{"contacted": contacted}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}



	updateTimeline = HistoryMessage{
		Title: "New Connection",
		Message: item.MyName + " Connected with you!",
		Timestamp: time.Now(),

	}

	connected := Connected{
		UserId: id,
		Name: item.MyName,
	}

	collection = client.Database(Database).Collection("history")
	filter = bson.M{"userid": item.UserId}
	update = bson.M{"$push": bson.M{"messages": updateTimeline}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	collection = client.Database(Database).Collection("follow")
	filter = bson.M{"userid": item.UserId}
	update = bson.M{"$push": bson.M{"connected": connected}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	}
}




func AddHistoryMessage(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request)  {
		vars := mux.Vars(r)
		id := vars["id"]
		var item HistoryMessage

		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("history")
		filter := bson.M{"userid": id}
		update :=  bson.M{"$push": bson.M{"messages": item}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Send a success response
		w.WriteHeader(http.StatusCreated)

	}
}

func StartConversation(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item Conversation

		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if a conversation with the same ByUserId and WithUserId exists
		collection := client.Database(Database).Collection("conversation")
		filter := bson.M{
			"$or": []bson.M{
				{"byuserid": item.ByUserId, "withuserid": item.WithUserId},
				{"byuserid": item.WithUserId, "withuserid": item.ByUserId},
			},
		}
		count, err := collection.CountDocuments(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if count > 0 {
			// Conversation already exists, return an error
			http.Error(w, "Conversation already exists", http.StatusFound)
			return
		}

		// Insert the item into the "conversation" collection in MongoDB
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteUserImages(minioClient *minio.Client, client *mongo.Client)http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request)  {
		vars := mux.Vars(r)
		
		id := vars["id"]

		var item DeleteImage
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		url := item.URL

		// Delete the image from the Minio bucket
		// Remove the object from Minio.
		err = minioClient.RemoveObject(context.Background(), "dumdum", url, minio.RemoveObjectOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete image from Minio: %v", err), http.StatusInternalServerError)
			return
		}

		// Update the "images" field of the corresponding MongoDB document
		collection := client.Database(Database).Collection("users")
		filter := bson.M{"images": url, "userid": id}
		update := bson.M{"$pull": bson.M{"images": url}}
		result, err := collection.UpdateMany(context.Background(), filter, update)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update MongoDB document: %v", err), http.StatusInternalServerError)
			return
		}
		if result.ModifiedCount == 0 {
			http.Error(w, fmt.Sprintf("No MongoDB documents updated"), http.StatusBadRequest)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)

	}
}

func UploadUserImages(minioClient *minio.Client, client *mongo.Client, minioURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request)  {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

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

			dotRemoved := extension[1:]

            // Generate a unique file name with the original extension.
            newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
			newPath := "https://"+minioURL + "/dumdum/" + newFilename
			ImagePaths = append(ImagePaths, newPath)

            // Upload the file to Minio.
            _, err = minioClient.PutObject(context.Background(), "dumdum", newFilename, file, fileHeader.Size, minio.PutObjectOptions{
				ContentType: "image/" + dotRemoved,
			})
            if err != nil {
				fmt.Println(err.Error())
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }

		fmt.Println(ImagePaths)

		collection := client.Database(Database).Collection("users")
		filter := bson.M{"userid": id}
		update := bson.M{"$push": bson.M{"images": bson.M{"$each": ImagePaths}}}
		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(result);

		updateTimeline := HistoryMessage{
			Title: "Uploaded image",
			Message: "Upload Image Success",
			Timestamp: time.Now(),

		}

		collection = client.Database(Database).Collection("history")
		filter = bson.M{"userid": id}
		update = bson.M{"$push": bson.M{"messages": updateTimeline}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}




		cursor, err := collection.Find(context.Background(), bson.M{"userid": id})
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

		fmt.Println(items)
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}


	}
}

func Upload(minioClient *minio.Client, minioURL string) http.HandlerFunc {
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

			dotRemoved := extension[1:]

			// Generate a unique file name with the original extension.
			newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
			newPath := "https://" + minioURL + "/dumdum/" + newFilename
			ImagePaths = append(ImagePaths, newPath)

			log.Println(ImagePaths)

			// Upload the file to Minio.
			_, err = minioClient.PutObject(context.Background(), "dumdum", newFilename, file, fileHeader.Size, minio.PutObjectOptions{
				ContentType: "image/" + dotRemoved,
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
func addRequirement(client *mongo.Client, FcmClient *messaging.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item Requirements
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item.Timestamp = time.Now();
		item.LastUpdate =  time.Now();
		
		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("requirements")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("devices")
		filter := bson.M{"userid": item.UserId}
		var userDevice UsersDevicesGet
		err = collection.FindOne(context.Background(), filter).Decode(&userDevice)
		if err != nil {
			log.Fatalf("Error finding user device: %v\n", err)
		}

		// Notification payload
		message := &messaging.Message{
			Notification: &messaging.Notification{
				Title: "Requirement Created",
				Body:  "Thanks For Creating a Requirement!",
			},
			Token: userDevice.DeviceId, // Replace with the FCM token of the specific device you want to send the notification to
		}
	
		// Send the notification
		_, err = FcmClient.Send(context.Background(), message)
		if err != nil {
			log.Fatalf("Error sending message: %v\n", err)
		}
	
		log.Println("Notification sent successfully.")
		
		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

func addDevice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into a UsersDevices struct
		var item UsersDevices
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Define a filter to find the existing document based on the UserId field
		filter := bson.M{"userid": item.UserId}

		// Define an update operation to either update the existing document or insert a new one
		update := bson.M{"$set": item}

		// Insert the item into the "devices" collection in MongoDB, but update if it already exists
		collection := client.Database(Database).Collection("devices")
		result, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if an update occurred (document already existed) or an insert occurred (new document)
		if result.UpsertedCount > 0 {
			// A new document was inserted
			w.WriteHeader(http.StatusCreated)
		} else {
			// The document was updated
			w.WriteHeader(http.StatusOK)
		}
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

		connections := Connections{
			UserId: item.UserId,
			Contacted: 0,
			Connections: 0,
		}
		

		collection = client.Database(Database).Collection("connections")
		_, err = collection.InsertOne(context.Background(), connections)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		history := History{
			UserId: item.UserId,
			Messages: []HistoryMessage{
				{Title: "Account Created", Message: "Hurray! you just crrated an account", Timestamp: time.Now()},
			},
		}

		collection = client.Database(Database).Collection("history")
		_, err = collection.InsertOne(context.Background(), history)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		follow := &Follow{
			UserId:    item.UserId,
			Contacted: []Contacted{},
			Connected: []Connected{},
		}
		collection = client.Database(Database).Collection("follow")
		_, err = collection.InsertOne(context.Background(), follow)
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
func editRequirement(client *mongo.Client) http.HandlerFunc {
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
		
		// Parse the request body into an Item struct
		var item RequirementsGet
		err = json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)

		item.LastUpdate =  time.Now();
		// tables := item.TableAttached
		
		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("requirements")

		filter := bson.M{"_id": oid}
		update := bson.M{"$set": item}
		_, err = collection.UpdateOne(context.Background(), filter, update)
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
		filter := bson.M{"userid": item.UserId}
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

func closeRequirement(client *mongo.Client) http.HandlerFunc {
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
		collection := client.Database(Database).Collection("requirements")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "closed", "lastUpdate": time.Now()}}
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
func getRequirements(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("requirements")
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		// Decode the cursor results into a slice of Item structs
		var items []RequirementsGet
		for cursor.Next(context.Background()) {
			var item RequirementsGet
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

func getFollowItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		// Get the item from the "follow" collection in MongoDB that matches the specified user ID
		collection := client.Database(Database).Collection("follow")
		var item Follow
		err := collection.FindOne(context.Background(), bson.M{"userid": id}).Decode(&item)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "No follow item found for the specified user ID", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(item)

		if len(item.Contacted) == 0 || len(item.Connected) == 0 {
			//http.Error(w, "Follow Null", http.StatusNotFound)
			log.Println("Follow Empty")
		}
		

		// Query the "users" collection to retrieve the user documents for the Contacted and Connected arrays
		usersColl := client.Database(Database).Collection("users")
		var contactedUsers []UserGet
		for _, contacted := range item.Contacted {
			var contactedUser UserGet
			err := usersColl.FindOne(context.Background(), bson.M{"userid": contacted.UserId}).Decode(&contactedUser)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contactedUsers = append(contactedUsers, contactedUser)
		}

		var connectedUsers []UserGet
		for _, connected := range item.Connected {
			var connectedUser UserGet
			err := usersColl.FindOne(context.Background(), bson.M{"userid": connected.UserId}).Decode(&connectedUser)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			connectedUsers = append(connectedUsers, connectedUser)
		}

		// Send the item as a JSON response with the retrieved user documents included
		itemWithUsers := struct {
			UserId     string     `json:"userid"`
			Contacted  []UserGet     `json:"contacted"`
			Connected  []UserGet     `json:"connected"`
		}{
			UserId:     item.UserId,
			Contacted:  contactedUsers,
			Connected:  connectedUsers,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(itemWithUsers)
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
		// oid, err := primitive.ObjectIDFromHex(id)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		cursor, err := collection.Find(context.Background(), bson.M{"userid": id})
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

		fmt.Println(items)
		
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
func getRequirement(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("requirements")
		
		cursor, err := collection.Find(context.Background(), bson.M{"userid": id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		
		// Decode the cursor results into a slice of Item structs
		var items []RequirementsGet
		for cursor.Next(context.Background()) {
			var item RequirementsGet

			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		fmt.Println(items)
		
		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}


func getDevice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("devices")
		
		cursor, err := collection.Find(context.Background(), bson.M{"userid": id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		
		// Decode the cursor results into a slice of Item structs
		var items []UsersDevicesGet
		for cursor.Next(context.Background()) {
			var item UsersDevicesGet

			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		fmt.Println(items)
		
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
		
		// Search for items in the "users" collection in MongoDB that match the search key in either the "name" or "company" field
		collection := client.Database(Database).Collection("users")
		filter := bson.M{
			"$or": []bson.M{
				{"name": primitive.Regex{Pattern: key, Options: "i"}},
				{"company": primitive.Regex{Pattern: key, Options: "i"}},
				{"busniess_category": primitive.Regex{Pattern: key, Options: "i"}},
			},
		}
		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		
		// Decode the cursor results into a slice of UserGet structs
		var items []UserGet
		for cursor.Next(context.Background()) {
			var item UserGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if item.BusniessType != "Individual" {
				items = append(items, item)
			}
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

