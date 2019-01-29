NINJAM DJ Bot

Install KX Studio: https://kxstudio.linuxaudio.org/Repositories

Install dependencies
```
sudo apt-get install libvorbis-dev sox libsox-fmt-mp3
sudo apt-get install x42-plugins calf-plugins liblilv-dev
```

Download and place to PATH (for exaple, copy to /usr/bin/):
```
https://sourceforge.net/projects/bs1770gain/files/bs1770gain/0.5.2/
```

Copy config.example.yaml to /etc/ninjam/djbot.yaml and configure it

Add to rc.local or any other autostart script

```
su dj -c '/usr/bin/ninjam-dj-bot -c /etc/ninjam/djbot.yaml' > /var/log/djbot.log 2>&1 &
```
