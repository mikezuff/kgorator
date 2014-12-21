kgorator is a cooling thermostat for your Raspberry PI. It's meant to be wired to control an outlet using a [relay](https://www.sparkfun.com/products/11042) wired to pin 17, and a 1-Wire temperature sensor [ds18b20](http://www.adafruit.com/products/381) wired to GPIO #4. Adafruit has a [good tutorial](https://learn.adafruit.com/adafruits-raspberry-pi-lesson-11-ds18b20-temperature-sensing/overview) on this.

kgorator supports
* configurable hysteresis
* persistent setpoints
* compressor recovery period

Simulation Mode
===============
Try it out in hardware simulation mode with the _-hwsim_ flag

Building for Raspberry Pi
=========================
[Cross compile](https://coderwall.com/p/pnfwxg) Go for RPi (ARM):
    cd $GOROOT/src
    GOOS=linux GOARCH=arm ./make.bash

Build kgorator for RPi
    cd $GOPATH/src/github.com/mikezuff/kgorator
    GOOS=linux GOARCH=arm go build ./cmds/kgorator

Auto Run
========
To make kgorator resilient to power failures and easy to use you can use it to replace the normal login on the default terminal. Instead of a login prompt you'll get kgorator. 

Install the binary
    sudo cp kgorator /usr/local/bin/kgorator
Add to /etc/modules
    w1-gpio
    w1-therm
In /etc/inittab, change the tty0 login line that was
    1:2345:respawn:/sbin/getty 38400 tty1
to be
    1:23:respawn:/sbin/getty -i -a pi -l /usr/local/bin/kgorator -o"" 38400 tty1

To upgrade kgorator you can't just copy the new binary to /usr/local/bin/kgorator, you'll get an error "Text file busy". If you go to runlevel 1 the init process will stop kgorator, then you can make the copy.
    init 1
    cp ~/kgorator /usr/local/bin
    init 2

Temperature Logger
==================
cmds/thermlog contains utility to log temperatures to a file
