---
title: Python image classification
author: GaborZeller
date: 2025-04-27T08-34-31Z
tags:
draft: true
---

# Python image classification

```python
import tensorflow as tf
from tensorflow.keras.preprocessing.image import ImageDataGenerator
from tensorflow.keras.models import Sequential
from tensorflow.keras.layers import Conv2D, MaxPooling2D, Flatten, Dense, Dropout
from tensorflow.keras.preprocessing import image
import numpy as np
import os

def create_model(img_height, img_width):
    model = Sequential([
        # Convolutional layers
        Conv2D(32, (3, 3), activation='relu', input_shape=(img_height, img_width, 3)),
        MaxPooling2D(2, 2),

        Conv2D(64, (3, 3), activation='relu'),
        MaxPooling2D(2, 2),

        Conv2D(64, (3, 3), activation='relu'),
        MaxPooling2D(2, 2),

        # Flatten and Dense layers
        Flatten(),
        Dense(64, activation='relu'),
        Dropout(0.5),
        Dense(1, activation='sigmoid')  # Binary classification (same/different)
    ])

    return model

def train_model():
    # Define parameters
    img_height = 150
    img_width = 150
    batch_size = 32
    epochs = 15

    # Define data directories
    train_dir = 'path/to/training/data'  # Your training data directory

    # Create data generators with augmentation
    train_datagen = ImageDataGenerator(
        rescale=1./255,
        rotation_range=20,
        width_shift_range=0.2,
        height_shift_range=0.2,
        shear_range=0.2,
        zoom_range=0.2,
        horizontal_flip=True,
        validation_split=0.2
    )

    # Load and prepare the training data
    train_generator = train_datagen.flow_from_directory(
        train_dir,
        target_size=(img_height, img_width),
        batch_size=batch_size,
        class_mode='binary',
        subset='training'
    )

    validation_generator = train_datagen.flow_from_directory(
        train_dir,
        target_size=(img_height, img_width),
        batch_size=batch_size,
        class_mode='binary',
        subset='validation'
    )

    # Create and compile the model
    model = create_model(img_height, img_width)
    model.compile(optimizer='adam',
                 loss='binary_crossentropy',
                 metrics=['accuracy'])

    # Train the model
    history = model.fit(
        train_generator,
        epochs=epochs,
        validation_data=validation_generator
    )

    # Save the model
    model.save('image_classifier_model.h5')

    return model

def predict_image(model, image_path):
    img_height = 150
    img_width = 150

    # Load and preprocess the image
    img = image.load_img(image_path, target_size=(img_height, img_width))
    img_array = image.img_to_array(img)
    img_array = np.expand_dims(img_array, axis=0)
    img_array /= 255.0

    # Make prediction
    prediction = model.predict(img_array)
    return prediction[0][0]

# Train the model
model = train_model()

# Example usage for prediction
def check_image(image_path, threshold=0.5):
    prediction = predict_image(model, image_path)
    is_similar = prediction >= threshold
    confidence = prediction if is_similar else 1 - prediction

    print(f"Prediction value: {prediction}")
    print(f"Is similar: {is_similar}")
    print(f"Confidence: {confidence * 100:.2f}%")

    return is_similar, confidence

# Example usage
test_image_path = 'path/to/test/image.jpg'
is_similar, confidence = check_image(test_image_path)
```

To use this code:

1. Set up your directory structure like this:

```
training_data/
    ├── class_1/
    │   ├── image1.jpg
    │   ├── image2.jpg
    │   └── ...
    └── class_2/
        ├── image1.jpg
        ├── image2.jpg
        └── ...
```

2. Install required packages:

```bash
pip install tensorflow pillow numpy
```

3. Modify the paths in the code:

- Change `train_dir` to point to your training data directory
- Change `test_image_path` to point to the image you want to test

4. Adjust the parameters if needed:

- `img_height` and `img_width`: image dimensions
- `batch_size`: number of images processed in each batch
- `epochs`: number of training iterations
- `threshold`: threshold for classification (default is 0.5)

Some important notes:

1. This example uses a CNN (Convolutional Neural Network) architecture suitable for image classification.

2. The model is trained for binary classification (similar/different), but you can modify it for multi-class classification by changing the last layer and loss function.

3. The training data should be organized in directories, where each directory name is the class label.

4. The model includes data augmentation to improve generalization.

5. You might need to adjust the model architecture and hyperparameters based on your specific use case.

6. For better results:
   - Use a larger dataset
   - Try different model architectures
   - Use transfer learning with pre-trained models like VGG16, ResNet, etc.
   - Experiment with different hyperparameters

Example usage:

```python
# Train the model
model = train_model()

# Test single image
test_image_path = 'path/to/test/image.jpg'
is_similar, confidence = check_image(test_image_path)

# Load saved model for later use
saved_model = tf.keras.models.load_model('image_classifier_model.h5')
is_similar, confidence = check_image(test_image_path, model=saved_model)
```

This is a basic example that you can build upon based on your specific requirements. For better results, you might want to consider using transfer learning with pre-trained models or more sophisticated architectures.
