// Import dependencies
const express = require('express');
const http = require('http');
const socketIO = require('socket.io');
const { MongoClient, ServerApiVersion } = require('mongodb');

// Create an Express app
const app = express();
const server = http.createServer(app);
const io = socketIO(server);

// Serve static files from the "public" directory
app.use(express.static('public'));

// MongoDB connection URL
const url = process.env.MONGO_URL;
const dbName = 'dumdum'; // Replace with your database name // Global variable for MongoDB client
let db;
// Connect to MongoDB
const client = new MongoClient(url, {
    serverApi: {
      version: ServerApiVersion.v1,
      strict: true,
      deprecationErrors: true,
    }
});

async function run() {
    try {
      // Connect the client to the server       (optional starting in v4.7)
      await client.connect();
      // Send a ping to confirm a successful connection
      await client.db(dbName).command({ ping: 1 });
      db = await client.db(dbName);
      console.log("Pinged your deployment. You successfully connected to MongoDB!");
    } catch (e){
        console.log(e)
    }
  }


// Handle incoming Socket.IO connections
io.on('connection', (socket) => {
  console.log('A user connected', socket.id);


  socket.on("leave", (roomID) => {
    console.log(`User Left Room: ${roomID}`);
    socket.leave(roomID);
  })

  // Handle user join events
  socket.on('join', (roomId) => {
    console.log(`User joined room ${roomId}`);
    socket.join(roomId);

    // Create a unique index on the roomID field
    const collection = db.collection('rooms');
    collection.createIndex({ roomID: 1 }, { unique: true })
      .then(() => {
        // Insert the room document if it doesn't exist
        return collection.findOneAndUpdate(
          { roomID: roomId },
          { $setOnInsert: { roomID: roomId, messages: [] } },
          { upsert: true, returnOriginal: false }
        );
      })
      .then((result) => {
        const room = result;
        console.log('Room document created or retrieved:', room);
      })
      .catch((err) => {
        console.error('Failed to insert or retrieve room:', err);
      });

    // Handle incoming messages
    socket.on('outgoing', (message) => {
        console.log('Received message:', message);
  
        // Check if the message with the same ID already exists
        collection.findOne({ roomID: roomId, 'messages.id': message.id })
          .then((existingMessage) => {
            if (existingMessage) {
              console.log('Message with the same ID already exists. Ignoring.');
              return;
            }
  
            // Broadcast the message to the other users in the room
            io.to(roomId).emit('incoming', message);
  
            // Update the room document with the new message
            collection.updateOne(
              { roomID: roomId },
              { $push: { messages: message } }
            )
              .then(() => {
                console.log('Room document updated with new message:', message);
              })
              .catch((err) => {
                console.error('Failed to update room:', err);
              });
          })
          .catch((err) => {
            console.error('Failed to check existing message:', err);
          });
      });
  });

  // Handle user disconnect events
  socket.on('disconnect', () => {
    console.log('A user disconnected');
  });
});

// HTTP route to get all messages in a room
app.get('/room/:roomId/messages', (req, res) => {
    const { roomId } = req.params;
    const collection = db.collection('rooms');
    
    collection.findOne({ roomID: roomId })
      .then((room) => {
        if (!room) {
          res.status(404).json({ error: 'Room not found' });
        } else {
          const messages = room.messages;
          res.json(messages);
        }
      })
      .catch((err) => {
        console.error('Failed to fetch room messages:', err);
        res.status(500).json({ error: 'Failed to fetch room messages' });
      });
});

const port = 3000;
server.listen(port, async () => {
    await run().catch(console.dir);
    console.log(`Server listening on port ${port}`);
});