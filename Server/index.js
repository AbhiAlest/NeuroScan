const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const handle = require('./handle');

const app = express();

// middleware
app.use(cors());
app.use(bodyParser.json());
app.use(handle);

// routes
app.post('/predict', async (req, res) => {
  // your predict route code here
});

// server
const PORT = process.env.PORT || 5000;
app.listen(PORT, () => console.log(`Server running on port ${PORT}`));
