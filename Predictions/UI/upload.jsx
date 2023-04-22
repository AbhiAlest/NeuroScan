import React, { useState } from 'react';
import axios from 'axios';

function App() {
  const [selectedFile, setSelectedFile] = useState(null);

  const fileChangedHandler = (event) => {
    setSelectedFile(event.target.files[0]);
  };

  const uploadHandler = () => {
    const formData = new FormData();
    formData.append('image', selectedFile);

    axios.post('/api/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }).then((response) => {
      console.log(response.data);
    });
  };

  return (
    <div>
      <h1>NeuroScan</h1>
      <input type="file" onChange={fileChangedHandler} />
      <button onClick={uploadHandler}>Upload</button>
    </div>
  );
}

export default App;
