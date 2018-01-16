package main

import (
	"testing"
	"reflect"
)

func TestDoAllDirs(t *testing.T) {
	got := doAllDirs([]string{"testdata/bad"})
	want := Reports{
		{36, "testdata/bad/gorutn_logimprt.go:5:\"log\" import is not allowed; instead use env.logger."},
		{99, "testdata/bad/gorutn_logimprt.go:13:go routines are not allowed inside adapters; instead use env.ScheduleWork or env.ScheduleDaemon."},
		{161, "testdata/bad/gorutn_logimprt2.go:5:\"log\" import is not allowed; instead use env.logger."},
		{226, "testdata/bad/gorutn_logimprt2.go:13:go routines are not allowed inside adapters; instead use env.ScheduleWork or env.ScheduleDaemon."},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("errors dont match\nwant:%v\ngot :%v", want, got)
	}
}


func TestDoAllDirsBadPath(t *testing.T) {
	// check no panics and no reports
	got := doAllDirs([]string{"testdata/unknown"})
	want := Reports{
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("errors dont match\nwant:%v\ngot :%v", want, got)
	}
}


func TestDoAllDirsGood(t *testing.T) {
	got := doAllDirs([]string{"testdata/good"})
	want := Reports{}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("errors dont match\nwant:%v\ngot :%v", want, got)
	}
}