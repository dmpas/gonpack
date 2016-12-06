package main

import "os"

func supplyDirectory(dirname string) {
	stat, err := os.Stat(dirname)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dirname, os.ModeDir|0777)
			if err == nil {
				return
			}
		}
		panic(err)
	}
	if !stat.IsDir() {
		panic("not a folder!!!")
	}
}
