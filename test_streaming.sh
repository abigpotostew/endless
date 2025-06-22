#!/bin/bash

# Test script for endless application
# This script helps debug deployment issues

echo "=== Testing Endless Application ==="

# Set the base URL (change this to your deployment URL)
BASE_URL="${1:-http://localhost:8080}"

echo "Testing base URL: $BASE_URL"

# Test 1: Health endpoint
echo -e "\n1. Testing health endpoint..."
curl -s -w "HTTP Status: %{http_code}\n" "$BASE_URL/health"

# Test 2: Homepage
echo -e "\n2. Testing homepage..."
curl -s -w "HTTP Status: %{http_code}\n" "$BASE_URL/" | head -20

# Test 3: Sitemap
echo -e "\n3. Testing sitemap..."
curl -s -w "HTTP Status: %{http_code}\n" "$BASE_URL/sitemap.xml" | head -10

# Test 4: Robots.txt
echo -e "\n4. Testing robots.txt..."
curl -s -w "HTTP Status: %{http_code}\n" "$BASE_URL/robots.txt"

# Test 5: Post page (this will likely fail without a trained model)
echo -e "\n5. Testing post page..."
curl -s -w "HTTP Status: %{http_code}\n" "$BASE_URL/post/123-test" | head -10

# Test 6: Train a model (if localhost)
if [[ "$BASE_URL" == *"localhost"* ]]; then
    echo -e "\n6. Testing model training..."
    echo "Training a simple model..."
    curl -s -X POST \
        -H "Content-Type: text/plain" \
        -d "This is a test story. It has multiple sentences. Each sentence should be processed by the Markov chain. The model will learn from this text and generate new stories." \
        -w "HTTP Status: %{http_code}\n" \
        "$BASE_URL/api/train"
fi

echo -e "\n=== Test Complete ===" 