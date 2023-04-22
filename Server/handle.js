import express from 'express';
import cors from 'cors';
import mongoose from 'mongoose';
import multer from 'multer';
import { createReadStream } from 'fs';
import brain from 'brain.js';
import path from 'path';

// Configure Multer for file upload
const upload = multer({ dest: 'uploads/' });

// Create a new Express application
const app = express();

// Enable CORS
app.use(cors());

// Connect to MongoDB
mongoose.connect('mongodb://localhost:27017/image-upload', {
  useNewUrlParser: true,
  useUnifiedTopology: true,
}).then(() => {
  console.log('Connected to MongoDB');
}).catch((error) => {
  console.error('Error connecting to MongoDB:', error);
});

// Define a schema for the image collection
const imageSchema = new mongoose.Schema({
  filename: { type: String, required: true },
  path: { type: String, required: true },
  mimetype: { type: String, required: true },
  size: { type: Number, required: true },
});

// Create a model for the image collection
const Image = mongoose.model('Image', imageSchema);

// Define a route for uploading images
app.post('/api/upload', upload.single('image'), async (req, res) => {
  try {
    // Get the uploaded file details
    const filename = req.file.filename;
    const mimetype = req.file.mimetype;
    const size = req.file.size;
    const path = req.file.path;

    // Create a new image document in the database
    const image = new Image({
      filename: filename,
      path: path,
      mimetype: mimetype,
      size: size,
    });
    await image.save();

    // Load the uploaded image and run it through the deep learning algorithm
    const net = new brain.recurrent.LSTM();
    const modelPath = path.join(__dirname, 'model.bin');
    const modelData = await createReadStream(modelPath).read();
    const model = JSON.parse(modelData);
    net.fromJSON(model);
    const imageStream = createReadStream(path);
    const chunks = [];
    imageStream.on('data', (chunk) => {
      chunks.push(chunk);
    });
    imageStream.on('end', () => {
      const buffer = Buffer.concat(chunks);
      const data = Array.from(buffer).map((x) => x / 255);
      const result = net.run(data);
      console.log('Brain.js prediction:', result);
    });

    // Send a response indicating that the image was uploaded successfully
    res.status(200).json({ message: 'Image uploaded successfully' });
  } catch (error) {
    console.error('Error uploading image:', error);
    res.status(500).json({ message: 'Error uploading image' });
  }
});

// Start the server
app.listen(8000, () => {
  console.log('Server running on port 8000');
});
