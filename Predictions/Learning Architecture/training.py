import numpy as np
from tensorflow import keras
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
from sklearn.utils import class_weight
from keras.preprocessing.image import load_img, img_to_array
from keras.layers.convolutional import Conv2D
from keras.layers.convolutional import MaxPooling2D
from keras.layers.core import Flatten
from keras.layers.core import Dense
from keras.layers.core import Dropout
from keras.layers.core import Reshape
from keras.layers.recurrent import LSTM
from keras.optimizers import Adam

# Load image data and labels
X = []
y = []
for tumor_type in ['meningioma', 'glioma', 'pituitary']:
    for i in range(1, 11):
        img_path = f'Brain Tumors (multiple)/{tumor_type}/{i}.png'
        img = load_img(img_path, target_size=(128, 128))
        img_array = img_to_array(img)
        X.append(img_array)
        y.append(tumor_type)
X = np.array(X)
y = np.array(y)

label_encoder = LabelEncoder()
integer_encoded = label_encoder.fit_transform(y)
onehot_encoder = OneHotEncoder(sparse=False)
integer_encoded = integer_encoded.reshape(len(integer_encoded), 1)
y = onehot_encoder.fit_transform(integer_encoded)

# Split data to training, validation, and test sets
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
X_train, X_val, y_train, y_val = train_test_split(X_train, y_train, test_size=0.2, random_state=42)

# Calculate class weights for dealing with handle imbalanced classes
class_weights = class_weight.compute_class_weight('balanced', np.unique(np.argmax(y_train, axis=1)), np.argmax(y_train, axis=1))
class_weights = dict(enumerate(class_weights))

# Hyperparameters
num_filters = 64
kernel_size = (3, 3)
activation = 'relu'
pool_size = (2, 2)
dropout_rate = 0.5
learning_rate = 0.0001
num_units = 64
num_epochs = 50
batch_size = 32
num_classes = len(np.unique(y))

# CNN model
cnn_model = keras.Sequential()
cnn_model.add(Conv2D(num_filters, kernel_size, activation=activation, input_shape=(128, 128, 3)))
cnn_model.add(MaxPooling2D(pool_size=pool_size))
cnn_model.add(Flatten())
cnn_model.add(Dense(128, activation=activation))
cnn_model.add(Dropout(dropout_rate))

# RNN model
rnn_model = keras.Sequential()
rnn_model.add(Reshape((-1, 128*128*3), input_shape=(X_train.shape[1], X_train.shape[2], X_train.shape[3])))
rnn_model.add(LSTM(num_units, return_sequences=True))
rnn_model.add(Flatten())
rnn_model.add(Dense(128, activation=activation))
rnn_model.add(Dropout(dropout_rate))

# Combine CNN and RNN models
combined_model = keras.Sequential()
combined_model.add(TimeDistributed(cnn_model, input_shape=(X_train.shape[1], X_train.shape[2], X_train.shape[3])))
combined_model.add(TimeDistributed(rnn_model))
combined_model.add(LSTM(num_units, return_sequences=False))

# Add dense layer and compile model
combined_model.add(Dense(num_classes, activation='softmax'))
optimizer = optimizers.Adam(lr=learning_rate)
combined_model.compile(loss='categorical_crossentropy', optimizer=optimizer, metrics=['accuracy'])

# Train the model
history = combined_model.fit(X_train, y_train, epochs=num_epochs, batch_size=batch_size, 
                              validation_data=(X_val, y_val), class_weight=class_weights)

# Evaluate the model
test_loss, test_acc = combined_model.evaluate(X_test, y_test, verbose=2)
print('Test accuracy:', test_acc)
