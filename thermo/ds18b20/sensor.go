package ds18b20

import (
	"bufio"
	"bytes"
	"fmt"
	"kgerator/thermo"
	"os"
	"strconv"
)

func Open(path string) (thermo.Meter, error) {
	d := ds18b20(path)
	return &d, nil
}

type ds18b20 string

func (d *ds18b20) String() string {
	temp, err := d.Sample()
	if err != nil {
		return fmt.Sprint(string(*d), "error reading:", err)
	}

	return fmt.Sprint(string(*d), " ", temp)
}

func (d *ds18b20) Sample() (thermo.F, error) {
	file, err := os.Open(string(*d))
	if err != nil {
		return 0, fmt.Errorf("Opening ds18b20 %s: %s", string(*d), err)
	}

	read := bufio.NewReader(file)
	line, err := read.ReadBytes('\n')
	if err != nil {
		return 0, fmt.Errorf("Reading ds18b20 %s: %s", string(*d), err)
	}

	if !bytes.HasSuffix(line, []byte("YES\n")) {
		return 0, fmt.Errorf("Reading ds18b20 %s: bad crc", string(*d))
	}

	line, err = read.ReadBytes('\n')
	if err != nil {
		return 0, fmt.Errorf("Reading temperature ds18b20 %s: %s", string(*d), err)
	}

	cidx := bytes.IndexByte(line, '=')
	if cidx == -1 {
		return 0, fmt.Errorf("Reading ds18b20 %s: bad format", string(*d))
	}

	n, err := strconv.ParseUint(string(bytes.TrimSpace(line[cidx+1:])), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Error parsing temperature ds18b20 %s: %s", string(*d), err)
	}

	c := thermo.C(float64(n) / 1000)
	return thermo.F(thermo.CtoF(c)), nil
}
