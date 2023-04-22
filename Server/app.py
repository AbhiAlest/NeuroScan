from flask import Flask, request, jsonify
import tensorflow as tf
import numpy as np
import cv2
from PIL import Image
from io import BytesIO
import base64

app = Flask(__name__)

# Load the trained model
model = tf.keras.models.load_model('trainedmodel.h5')

@app.route('/api/upload', methods=['POST'])
def upload():
    file = request.files['image']
    img_str = file.read()

    # Decode the image and convert it to grayscale
    img = Image.open(BytesIO(img_str))
    img = np.array(img)
    img = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    img = cv2.resize(img, (128, 128))
    img = np.expand_dims(img, axis=-1)
    img = np.expand_dims(img, axis=0)

    # Run the image through the trained model and get the prediction
    pred = model.predict(img)
    if pred > 0.5:
        result = 'Cancerous'
    else:
        result = 'Non-Cancerous'

    return jsonify({'result': result})
