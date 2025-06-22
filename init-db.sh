#!/bin/bash

# Database initialization script for endless application
# This script trains a model to prevent 404 errors on post pages

echo "=== Initializing Endless Database ==="

# Set the base URL (default to localhost, can be overridden)
BASE_URL="${1:-http://localhost:8080}"

echo "Initializing database at: $BASE_URL"

# Sample training text with multiple sentences
TRAINING_TEXT="The quick brown fox jumps over the lazy dog. This is a sample story for training the Markov chain model. The model will learn from this text and generate new stories. Each sentence should end with proper punctuation. This ensures the model works correctly. Stories are generated based on patterns learned from training data. The system creates unique narratives for each request. Users can explore different stories by visiting various post URLs. The application provides an endless stream of creative content. Markov chains are powerful tools for text generation. They analyze word patterns and create coherent sentences. This technology enables automated storytelling at scale."

echo "Training model with sample text..."

# Train the model
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: text/plain" \
  -d "$TRAINING_TEXT" \
  -w "HTTP Status: %{http_code}" \
  "$BASE_URL/api/train")

echo "Response: $RESPONSE"

# Check if training was successful
if [[ "$RESPONSE" == *"HTTP Status: 201"* ]]; then
    echo "✅ Model trained successfully!"
    echo "✅ Database initialized!"
    echo ""
    echo "You can now test the application:"
    echo "  Homepage: $BASE_URL/"
    echo "  Sample post: $BASE_URL/post/123-test"
else
    echo "❌ Failed to train model"
    echo "Response: $RESPONSE"
    echo ""
    echo "Troubleshooting:"
    echo "1. Make sure the application is running"
    echo "2. Check if the API endpoint is accessible"
    echo "3. Verify network connectivity"
    echo "4. Check application logs for errors"
fi

echo ""
echo "=== Initialization Complete ===" 