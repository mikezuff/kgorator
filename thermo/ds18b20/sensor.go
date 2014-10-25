// ds18b20 implements kgerator/thermo.Meter for the ds18b20 1-Wire thermometer
// http://datasheets.maximintegrated.com/en/ds/DS18B20.pdf
package ds18b20

import (
	"bufio"
	"bytes"
	"fmt"
	"kgerator/thermo"
	"os"
	"strconv"
)

// Open returns a thermo.Meter using the ds18b20 device at path.
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
		return 0, fmt.Errorf("Opening ds18b20: %s", err)
	}

	read := bufio.NewReader(file)
	line, err := read.ReadBytes('\n')
	if err != nil {
		return 0, fmt.Errorf("Reading ds18b20: %s", err)
	}

	if !bytes.HasSuffix(line, []byte("YES\n")) {
		return 0, fmt.Errorf("Reading ds18b20: bad crc")
	}

	line, err = read.ReadBytes('\n')
	if err != nil {
		return 0, fmt.Errorf("Reading temperature ds18b20: %s", err)
	}

	cidx := bytes.IndexByte(line, '=')
	if cidx == -1 {
		return 0, fmt.Errorf("Reading ds18b20: bad format")
	}

	n, err := strconv.ParseUint(string(bytes.TrimSpace(line[cidx+1:])), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Error parsing temperature ds18b20: %s", err)
	}

	c := thermo.C(float64(n) / 1000)
	return thermo.F(thermo.CtoF(c)), nil
}
