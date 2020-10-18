package readini

import "testing"

func Testread(t *testing.T) {
	c := new(config)
	c.filePath = "test.ini"
	c.readini()
	section := "paths"
	key := "data"
	value := "/home/git/grafana"
	k, ok1 := c.info[section]
	v, ok2 := c.info[section][key]
	if v != value || !ok1 || !ok2 {
		t.Errorf("expected: section : " + section + "; key : " + key + "; value : " + value)
		t.Errorf("but got: section : " + section + "; key : " + k[key] + "; value : " + v)
	}
}
