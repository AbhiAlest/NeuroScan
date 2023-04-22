import React, { useState } from 'react';
import axios from 'axios';

interface AppProps {}

interface AppState {
  selectedFile: File | null;
}

class App extends React.Component<AppProps, AppState> {
  constructor(props: AppProps) {
    super(props);
    this.state = {
      selectedFile: null,
    };
  }

  fileChangedHandler = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      this.setState({ selectedFile: event.target.files[0] });
    }
  };

  uploadHandler = () => {
    if (this.state.selectedFile) {
      const formData = new FormData();
      formData.append('image', this.state.selectedFile);

      axios.post('/api/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }).then((response) => {
        console.log(response.data);
      });
    }
  };

  render() {
    return (
      <div>
        <h1>NeuroScan</h1>
        <input type="file" onChange={this.fileChangedHandler} />
        <button onClick={this.uploadHandler}>Upload</button>
      </div>
    );
  }
}

export default App;
