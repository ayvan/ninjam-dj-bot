**POST /v1/tracks/**

HTTP codes:
201
400
404

Simple multipart/form-data file upload, .mp3 only. Form field name "file".
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

Для установки тегов трека следует в поле tags поместить массив из объектов с ID тега, прочие поля объектов игнорируются:
```json
[
  {
    "id":1
  },
  {
    "id":3
  }
]
```

Все теги трека перезаписываются теми, которые были переданы в массиве "tags", таким образом, управление тегами
сводится к редактированию этого массива.

HTTP codes:
200
400
404

Example request:
```json
{
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
  "author_id":1,
  "loop_start": 1827878,
  "loop_end": 16373318,
  "bpm": 132,
  "bpi": 16,
  "key": 7,
  "mode": 1
}
```

**POST /v1/authors/**

HTTP codes:
201
400

Example request:
```json
{
  "name":"Burillo",
  "url":"",
  "description":""
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

**GET /v1/authors/:id**

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

**PUT /v1/authors/:id**

HTTP codes:
200
400
404

Example request:
```json
{
  "name":"Burillo",
  "url":"",
  "description":""  
}
```

**POST /v1/playlists/**

Треки плейлиста содержатся в поле "tracks" в виде массива объектов:
```json
{
  "track_id":1, // ID трека
  "repeats":10, // число повторов трека при его воспроизведении
  "timeout":60, // таймаут после окончания трека и перед воспроизведением следующего
  "queue":true // активна ли очередь музыкантов, т.е. будет ли объявляться, кто играет следующим по очереди
}
```

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

Поле "target_track_time" содержит "шаблонное" время трека в плейлисте, в секундах. Когда в существующий плейлист
добавляется трек, можно автоматически рассчитать число повторов, взяв target_track_time из плейлиста и подогнав
время трека (с повторами) под это время.

HTTP codes:
200
400

Example response:
````json
[{
  "id": 4,
  "name": "test 2",
  "description": "",
  "target_track_time": 0,
  "tracks": [
    {
      "track_id": 1,
      "repeats": 10,
      "timeout": 60,
      "queue": true
    },
    {
      "track_id": 2,
      "repeats": 15,
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
  "target_track_time": 0,
  "tracks": [
    {
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
  "target_track_time": 0,
  "tracks": [
    {
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
