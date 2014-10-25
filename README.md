Building for Raspberry Pi
=========================
Build go for RPi: GOOS=linux GOARCH=arm
GOOS=linux GOARCH=arm ./make.bash
https://coderwall.com/p/pnfwxg

Build kgorator for RPi
GOOS=linux GOARCH=arm go build kgorator.go

Autorun
=======
To make kgorator start when the Raspberry Pi boots:

Add w1-gpio and w1-therm to /etc/modules
In /etc/inittab, change the tty0 login line that was
1:2345:respawn:/sbin/getty 38400 tty1
to be
1:23:respawn:/sbin/getty -i -a root -l /home/pi/kgorator-install/kgorator -o"" 38400 tty1

mkdir /home/pi/kgorator-install
cp kgorator /home/pi/kgorator-install/kgorator-vX
ln -s /home/pi/kgorator-install/kgorator-vX /home/pi/kgorator-install/kgorator


