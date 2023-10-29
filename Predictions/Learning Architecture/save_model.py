# saves model into .h5 file.

from tensorflow import keras

# Load the trained model (assumes you have already trained it)
trained_model = keras.models.load_model('brain_tumor_detection_model.h5')

# Save the trained model to a file
trained_model.save('brain_tumor_detection_model_final.h5')
