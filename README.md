# GO backend services for react-counter [Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
This repo consists of two go microservices and the docker-compose file for an MQTT server. Button presses from the react-native mobile app [repo here](https://github.com/jhawk7/react-counter) are sent to the backend server that emits events to the MQTT server. Those events are consumed and stored in Redis and persisted via a Append-Only-File (AOF).
- `Producer` - receives events via persisted websocket connection from react native app and forwards them to an MQTT Server
- `MQTT Server` - brokers events; stores until consumed; this would usually be done with kafka in a prodcution system handling a ton of events
- `Consumer` - an mqtt client that consumes the events and stores the corresponding count in a Redis instance
- `Redis` - INCR/DECR count according to event and updates AOF; will replay AOF events on restart if it goes down

- This is a rather complicated setup for simply keeping track of a button press count, but the goal here was to emulate an actual system for processing and storing potentiall many concurrent events in real time.
