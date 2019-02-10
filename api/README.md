**POST /v1/tracks/**

HTTP codes:
201
400
404

Simple form file upload, .mp3 only.
Response JSON same as GET request response.

**GET /v1/tracks/**

HTTP codes:
200
400
404

Example response:
```json
[{
  "id": 1,
  "file_path": "DrumLoop.mp3",
  "title": "Drum Loop",
  "artist": "Burillo",
  "album": "Best of PornHub",
  "album_track_number": 1,
  "played": 0,
  "author_id":1,
  "author":{
    "id":1,
    "name":"Burillo",
    "url":"",
    "description":""
  },
  "length": 0,
  "loop_start": 1827878,
  "loop_end": 16373318,
  "bpm": 132,
  "bpi": 16,
  "key": 7,
  "mode": 1,
  "integrated": -14.1,
  "range": 5.14,
  "peak": -0.42,
  "shortterm": -13.83,
  "momentary": -11.74
}]
```

**GET /v1/tracks/:id**

HTTP codes:
200
400
404

Example response:
```json
{
  "id": 1,
  "file_path": "DrumLoop.mp3",
  "title": "Drum Loop",
  "artist": "Burillo",
  "album": "Best of PornHub",
  "album_track_number": 1,
  "played": 0,
  "author_id": 1,
  "author":{
    "id":1,
    "name":"Burillo",
    "url":"",
    "description":""
  },
  "length": 0,
  "loop_start": 1827878,
  "loop_end": 16373318,
  "bpm": 132,
  "bpi": 16,
  "key": 7,
  "mode": 1,
  "integrated": -14.1,
  "range": 5.14,
  "peak": -0.42,
  "shortterm": -13.83,
  "momentary": -11.74
}
```

**PUT /v1/tracks/:id**

HTTP codes:
200
400
404

Example request:
```json
{
  "file_path": "DrumLoop.mp3",
  "title": "Drum Loop",
  "artist": "Burillo",
  "album": "Best of PornHub",
  "album_track_number": 1,
  "tags": [
    {
      "id":1
    } ,
    {
      "id":3
    }
  ],
  "author_id":1
}
```

**POST /v1/authors/**

HTTP codes:
201
400

Example request:
```json
{
  "name":"Burillo"
}
```

**GET /v1/authors/**

HTTP codes:
200
400
404

Example response:
```json
[{
  "id": 1,
  "name":"Burillo",
  "url":"",
  "description":""
}]
```

**GET /v1/tracks/:id**

HTTP codes:
200
400
404

Example response:
```json
{
  "id": 1,
  "name":"Burillo",
  "url":"",
  "description":""
}
```

**PUT /v1/tracks/:id**

HTTP codes:
200
400
404

Example request:
```json
{
  "id": 1,
  "name":"Burillo",
  "url":"",
  "description":""  
}
```

**POST /v1/playlists/**

HTTP codes:
200
400

Example request:
```json
{
  "name":"Playlist 1",
  "tracks":[
    {
      "track_id":1,
      "repeats":10,
      "timeout":60,
      "queue":true
    },
    {
      "track_id":1,
      "repeats":5,
      "timeout":120,
      "queue":true
    }
  ]
}
```

**GET /v1/playlists/**

HTTP codes:
200
400

Example response:
````json
[{
  "id": 4,
  "name": "test 2",
  "description": "",
  "track_time": 0,
  "tracks": [
    {
      "id": 4,
      "playlist_id": 4,
      "track_id": 1,
      "repeats": 10,
      "timeout": 60,
      "queue": true
    }
  ]
}]

````

**GET /v1/playlists/:id**

HTTP codes:
200
400
404


Example response:
````json
{
  "id": 4,
  "name": "Playlist 4",
  "description": "",
  "track_time": 0,
  "tracks": [
    {
      "id": 4,
      "playlist_id": 4,
      "track_id": 1,
      "repeats": 10,
      "timeout": 60,
      "queue": true
    }
  ]
}

````

**PUT /v1/playlists/:id**

HTTP codes:
200
400
404

Example request:
```json
{
  "name": "My New Playlist",
  "description": "",
  "track_time": 0,
  "tracks": [
    {
      "id": 4,
      "track_id":1,
      "repeats": 10,
      "timeout": 80,
      "queue": true
    }
  ]
}
```

**GET /v1/tags**

HTTP codes:
200
400

Example response:
```json
[
  {
    "id": 1,
    "name": "Metal"
  },
  {
    "id": 2,
    "name": "Rock"
  },
  {
    "id": 3,
    "name": "Blues"
  },
  {
    "id": 4,
    "name": "Hell"
  }
]
```

**GET /v1/tags/:id**

HTTP codes:
200
400
404

Example response:
```json
{
  "id": 1,
  "name": "Metal"
}
```

**POST /v1/tags/**

HTTP codes:
200
400

Example request:
```json
{
  "name":"Death"
}
```

**PUT /v1/tags/:id**

HTTP codes:
200
400
404

Example request:
```json
{
  "name":"Rock"
}
```
