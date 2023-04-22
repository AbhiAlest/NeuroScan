const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const { predict } = require('./predict');

const app = express();

// middleware
app.use(cors());
app.use(bodyParser.json());

// routes
app.post('/predict', async (req, res) => {
  try {
    const result = await predict(req.body.image);
    res.status(200).json(result);
  } catch (error) {
    console.log(error);
    res.status(500).json({ message: 'Something went wrong.' });
  }
});

// server
const PORT = process.env.PORT || 5000;
app.listen(PORT, () => console.log(`Server running on port ${PORT}`));
