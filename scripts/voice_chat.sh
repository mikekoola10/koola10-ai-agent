#!/bin/bash

# Default endpoint
ENDPOINT="https://koola10.fly.dev/voice/input"

if [ "$1" != "" ]; then
    ENDPOINT=$1
fi

echo "Recording... Press Ctrl+C to stop."
# Record to a temporary wav file
temp_file=$(mktemp).wav
# Use sox or ffmpeg to record. Ffmpeg is common.
# -t 5 for 5 seconds, or let user Ctrl+C
ffmpeg -f alsa -i default -t 10 "$temp_file" -y -loglevel quiet

echo "Transcribing and sending to agent..."
response=$(curl -s -X POST -F "file=@$temp_file" "$ENDPOINT")

echo "Agent Response:"
echo "$response" | jq .

audio_url=$(echo "$response" | jq -r .audio_url)
if [ "$audio_url" != "null" ]; then
    echo "Playing response audio..."
    # You can use ffplay or vlc to play the audio url
    ffplay -nodisp -autoexit "$audio_url" >/dev/null 2>&1
fi

rm "$temp_file"
