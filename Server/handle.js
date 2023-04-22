const express = require('express');
const multer = require('multer');
const axios = require('axios');

const app = express();

// Multer config
const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, 'uploads/');
  },
  filename: (req, file, cb) => {
    cb(null, file.originalname);
  }
});

const upload = multer({ storage });

// handle file uploads
app.post('/upload', upload.single('mriImage'), async (req, res) => {
  try {
    const { filename, path } = req.file;

    // Send file to Go server for processing
    const response = await axios.post('http://localhost:8080/process', {
      filename, //do this later
      path,
    });

    // Save response to MongoDB
    // ...

    res.status(200).json({
      message: 'File uploaded and processed successfully',
      data: response.data,
    });
  } catch (error) {
    console.error(error);
    res.status(500).json({ error: 'An error occurred while processing the file.' });
  }
});

app.listen(3000, () => {
  console.log('Server started on port 3000');
});
