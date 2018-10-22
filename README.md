NINJAM DJ Bot

Install pulseaudio and mpg321
```
sudo apt-get install pulseaudio
sudo apt-get install mpg321
```

Test pulseaudio
```
pulseaudio --start -D

cat .asoundrc
```

Must print
```
pcm.pulse { type pulse }
ctl.pulse { type pulse }
pcm.!default { type pulse }
ctl.!default { type pulse }
```

Compile cninjam (ninjam cursesclient) from https://github.com/justinfrankel/ninjam (ninjam/cursesclient/) and then start it:
```
./cninjam guitar-jam.ru:2051 -audiostr "in pulse out null" -user anonymous:test -sessiondir /dev/null
```

To autostart, in Ubuntu you can create /etc/supervisor/conf.d/cninjam.conf
```
[program:cninjam]
command=/usr/bin/cninjam guitar-jam.ru:2051 -audiostr "in pulse out null" -user jamtrack -pass DJBOTPASS -nosavesourcefiles -sessiondir /dev/null
stdout_logfile=/var/log/cninjam.log
autostart=true
autorestart=true
user=dj
stopsignal=KILL
numprocs=1
```


Copy config.example.yaml to /etc/ninjam/djbot.yaml and configure it

Add to rc.local or any other autostart script

```
su dj -c '/usr/bin/ninjam-dj-bot -c /etc/ninjam/djbot.yaml' > /var/log/djbot.log 2>&1 &

su dj -c 'pulseaudio --start -D'
```
