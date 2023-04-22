import os
import tensorflow as tf
from tensorflow.keras.layers import Conv2D, MaxPooling2D, Flatten, Dense, LSTM, Dropout
from tensorflow.keras.models import Sequential
from tensorflow.keras.optimizers import Adam
from tensorflow.keras.preprocessing.image import ImageDataGenerator
from sklearn.model_selection import train_test_split, GridSearchCV, RandomizedSearchCV
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
import numpy as np
import pandas as pd
import random

# random seed for reproducibility
random.seed(42)
np.random.seed(42)
tf.random.set_seed(42)

brain_tumor_dir = "/path/to/brain_tumor_dataset"
glisoma_dir = "/path/to/glisoma_dataset"

# grid search hyperparameters
cnn_param_grid = {
    'num_filters': [16, 32, 64],
    'kernel_size': [(3,3), (5,5), (7,7)],
    'activation': ['relu', 'sigmoid'],
    'pool_size': [(2,2), (3,3)],
    'dropout_rate': [0.2, 0.3, 0.4],
    'learning_rate': [1e-4, 1e-3, 1e-2]
}

rnn_param_grid = {
    'num_layers': [1, 2, 3],
    'num_neurons': [32, 64, 128],
    'dropout_rate': [0.2, 0.3, 0.4],
    'learning_rate': [1e-4, 1e-3, 1e-2]
}

# Load datasets
def load_data():
    glioma_path = os.path.join(glisoma_dir, "kaggle_3m")
    glioma_df = pd.read_csv(os.path.join(glioma_path, "data.csv"))
    glioma_images_path = os.path.join(glioma_path, "images")
    glioma_masks_path = os.path.join(glioma_path, "masks")

    train_dir = os.path.join(brain_tumor_dir, "Brain Tumors (multiple)")

    # ImageDataGenerator for editing images
    datagen = ImageDataGenerator(rescale=1./255,
                                 rotation_range=20,
                                 width_shift_range=0.2,
                                 height_shift_range=0.2,
                                 shear_range=0.2,
                                 zoom_range=0.2,
                                 horizontal_flip=True,
                                 vertical_flip=True,
                                 validation_split=0.2)
    
    # Load and split training data to training and validation sets
    train_generator = datagen.flow_from_directory(train_dir,
                                                  target_size=(256, 256),
                                                  batch_size=32,
                                                  class_mode='categorical',
                                                  subset='training',
                                                  shuffle=True)
    
    val_generator = datagen.flow_from_directory(train_dir,
                                                target_size=(256, 256),
                                                batch_size=32,
                                                class_mode='categorical',
                                                subset='validation',
                                                shuffle=True)
    
    return train_generator, val_generator, glioma_df, glioma_images_path, glioma_masks_path

# CNN model 
def create_cnn_model(num_filters, kernel_size, activation, pool_size, dropout_rate):
    model = keras.Sequential()
    model.add(keras.layers.Conv2D(num_filters, kernel_size, activation=activation, input_shape=(128, 128, 1)))
    model.add(keras.layers.MaxPooling2D(pool_size=pool_size))
    model.add(keras.layers.Flatten())
    model.add(keras.layers.Dropout(dropout_rate))
    model.add(keras.layers.Dense(3, activation='softmax'))
    return model


# RNN model
def create_rnn_model(num_units, dropout_rate, learning_rate):
    model = keras.Sequential()
    model.add(keras.layers.LSTM(num_units, input_shape=(X_train.shape[1], X_train.shape[2])))
    model.add(keras.layers.Dropout(dropout_rate))
    model.add(keras.layers.Dense(2, activation='softmax'))
    optimizer = keras.optimizers.Adam(learning_rate=learning_rate)
    model.compile(optimizer=optimizer, loss='binary_crossentropy', metrics=['accuracy'])
    return model


# Build and train CNN model
def build_cnn_model(num_filters, kernel_size, activation, pool_size, dropout_rate, learning_rate):
    model = create_cnn_model(num_filters, kernel_size, activation, pool_size, dropout_rate)
    optimizer = keras.optimizers.Adam(learning_rate=learning_rate)
    model.compile(optimizer=optimizer, loss='categorical_crossentropy', metrics=['accuracy'])
    model.fit(train_data, epochs=10)
    return model


# Build and train RNN model
def build_rnn_model(num_units, dropout_rate, learning_rate):
    model = create_rnn_model(num_units, dropout_rate, learning_rate)
    model.fit(X_train, y_train, epochs=10, validation_data=(X_test, y_test))
    return model


# Define hyperparameter grid for CNN model
cnn_param_grid = {
    'num_filters': [32, 64, 128],
    'kernel_size': [(3,3), (5,5), (7,7)],
    'activation': ['relu', 'tanh', 'sigmoid'],
    'pool_size': [(2,2), (3,3), (4,4)],
    'dropout_rate': [0.2, 0.3, 0.4],
    'learning_rate': [0.001, 0.01, 0.1]
}

# Define hyperparameter grid for RNN model
rnn_param_grid = {
    'num_units': [32, 64, 128],
    'dropout_rate': [0.2, 0.3, 0.4],
    'learning_rate': [0.001, 0.01, 0.1]
}

# grid search for CNN model
cnn_model = keras.wrappers.scikit_learn.KerasClassifier(build_cnn_model)
cnn_grid_search = GridSearchCV(cnn_model, param_grid=cnn_param_grid, cv=3)
cnn_grid_search.fit(train_data)

# grid search for RNN model
X, y, _ = load_data()
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
rnn_model = keras.wrappers.scikit_learn.KerasClassifier(build_rnn_model)
rnn_grid_search = GridSearchCV(rnn_model, param_grid=rnn_param_grid, cv=3)
rnn_grid_search.fit(X_train, y_train)

# Print the best hyperparameters and accuracy scores for CNN and RNN models
print("CNN Test Accuracy:", cnn_test_acc)
print("RNN Test Accuracy:", rnn_test_acc)
