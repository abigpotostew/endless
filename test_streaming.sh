#!/bin/bash

echo "Testing the streaming page functionality..."
echo ""

# Start the server in the background
echo "Starting server..."
./endless &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Server started with PID: $SERVER_PID"
echo ""

# Test the streaming endpoint
echo "Testing streaming page at /post/42/stream"
echo "This will take about 8 seconds to complete with natural jitter..."
echo ""

# Use curl to test the streaming endpoint
curl -N "http://localhost:8080/post/42/stream"

echo ""
echo ""
echo "Streaming test completed!"
echo ""

# Stop the server
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Server stopped."
echo ""
echo "To test manually:"
echo "1. Start the server: ./endless"
echo "2. Visit: http://localhost:8080/post/42/stream"
echo "3. Watch the content appear gradually over ~8 seconds with natural jitter" 